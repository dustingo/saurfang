package taskhandler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/models/notify"
	"saurfang/internal/models/task"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/ntfy"
	"saurfang/internal/tools/pkg"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gofiber/fiber/v3"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	nomadapi "github.com/hashicorp/nomad/api"
)

type CustomTaskHandler struct {
	base.BaseGormRepository[task.CustomTask]
	NomadClient *nomadapi.Client
	Monitor     *pkg.NomadMonitor
}

func NewCustomTaskHandler() *CustomTaskHandler {
	monitor, err := pkg.NewNomadMonitor()
	if err != nil {
		slog.Error("failed to create nomad monitor", "error", err)
		os.Exit(1)
	}

	return &CustomTaskHandler{
		BaseGormRepository: base.BaseGormRepository[task.CustomTask]{DB: config.DB},
		NomadClient:        config.NomadCli,
		Monitor:            monitor,
	}
}

// Handler_CreateCustomTask 创建自定义任务
func (h *CustomTaskHandler) Handler_CreateCustomTask(c fiber.Ctx) error {
	var payload task.CustomTaskPayload
	if err := c.Bind().Body(&payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid request body", err.Error(), nil)
	}

	// 验证脚本类型
	if !h.isValidScriptType(payload.ScriptType) {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid script type", "unsupported script type", nil)
	}

	customTask := task.CustomTask{
		Name:        payload.Name,
		Description: payload.Description,
		ScriptType:  payload.ScriptType,
		Script:      payload.Script,
		TargetHosts: payload.TargetHosts,
		Parameters:  payload.Parameters, // 直接使用字符串，因为已经是 JSON 字符串格式
		Status:      "active",
		Timeout:     payload.Timeout,
		RetryCount:  payload.RetryCount,
	}

	if err := h.Create(&customTask); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create custom task", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "custom task created successfully", "", customTask)
}

// Handler_ExecuteCustomTask 立即执行自定义任务
func (h *CustomTaskHandler) Handler_ExecuteCustomTask(c fiber.Ctx) error {
	taskID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid task id", err.Error(), nil)
	}

	var customTask task.CustomTask
	result, err := h.ListByID(uint(taskID))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusNotFound, 1, "task not found", err.Error(), nil)
	}
	customTask = *result

	// 创建执行记录
	execution := task.CustomTaskExecution{
		TaskID:        uint(taskID),
		Status:        "pending",
		StartTime:     time.Now(),
		CheckInterval: 10, // 10秒检查一次
		MaxCheckCount: 60, // 最多检查60次（10分钟）
	}

	if err := config.DB.Create(&execution).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create execution record", err.Error(), nil)
	}

	// 异步执行任务
	go h.ExecuteCustomTaskAsync(&customTask, &execution)

	// 开始监控任务执行状态
	h.Monitor.StartMonitoring(&execution)

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "custom task execution started", "", fiber.Map{
		"execution_id": execution.ID,
		"task_id":      taskID,
	})
}

// executeCustomTaskAsync 异步执行自定义任务
func (h *CustomTaskHandler) ExecuteCustomTaskAsync(customTask *task.CustomTask, execution *task.CustomTaskExecution) {
	defer func() {
		execution.EndTime = &time.Time{}
		*execution.EndTime = time.Now()
		config.DB.Model(execution).Updates(map[string]interface{}{
			"status":   execution.Status,
			"end_time": execution.EndTime,
		})

	}()
	var successCount, failCount int
	var successJobs, failedJobs []string
	var mu sync.Mutex

	// 解析目标主机
	targetHosts := h.parseTargetHosts(customTask.TargetHosts)
	if len(targetHosts) == 0 {
		execution.Status = "failed"
		execution.ErrorMsg = "no target hosts specified"
		return
	}

	// 为每个目标主机生成单独的Job配置
	var dispatchResults []string
	var jobIDs []string

	for _, host := range targetHosts {
		// 为每个主机创建任务副本，更新target_hosts参数
		hostTask := *customTask
		hostTask.TargetHosts = host // 单个主机名

		// 生成Nomad Job配置
		jobSpec, err := h.generateNomadJobSpec(&hostTask, host)
		if err != nil {
			slog.Error("Failed to generate job spec for host", "host", host, "error", err)
			dispatchResults = append(dispatchResults, fmt.Sprintf("Failed to generate job spec for %s: %v", host, err))
			h.recordFailedJob(&mu, &failCount, &failedJobs, host)
			continue
		}

		// 保存Job HCL到Consul（使用主机特定的key）
		consulKey := fmt.Sprintf("custom_job/%d_%s", customTask.ID, host)
		if err = h.saveJobHCLToConsulWithKey(&hostTask, jobSpec, consulKey); err != nil {
			slog.Warn("Failed to save job HCL to Consul", "host", host, "error", err)
			// 不阻止任务执行，只记录警告
		}

		// 解析并注册Nomad Job
		job, err := h.NomadClient.Jobs().ParseHCL(jobSpec, true)
		if err != nil {
			slog.Error("Failed to parse job spec for host", "host", host, "error", err)
			dispatchResults = append(dispatchResults, fmt.Sprintf("Failed to parse job spec for %s: %v", host, err))
			h.recordFailedJob(&mu, &failCount, &failedJobs, host)
			continue
		}

		// 使用Nomad自动生成的job.ID，不手动设置
		// job.ID 由Nomad根据job名称自动生成

		// 注册Job
		resp, _, err := h.NomadClient.Jobs().Register(job, &nomadapi.WriteOptions{})
		if err != nil {
			slog.Error("Failed to register job for host", "host", host, "error", err)
			dispatchResults = append(dispatchResults, fmt.Sprintf("Failed to register job for %s: %v", host, err))
			h.recordFailedJob(&mu, &failCount, &failedJobs, host)

		} else {
			// 使用Nomad返回的job ID
			jobIDs = append(jobIDs, *job.ID)
			dispatchResults = append(dispatchResults, fmt.Sprintf("Registered job for %s: %s (eval: %s)", host, *job.ID, resp.EvalID))
			h.recordSuccessJob(&mu, &successCount, &successJobs, host)
		}
	}
	ntfy.PublishNotification(notify.EventTypeCustomJob, fmt.Sprintf("custom task %s", customTask.Name), successJobs, failedJobs, successCount, failCount)

	// 存储所有Job ID，用逗号分隔
	execution.NomadJobID = strings.Join(jobIDs, ",")
	execution.Status = "running"
	execution.Result = fmt.Sprintf("Jobs registered for all hosts. Results: %s", strings.Join(dispatchResults, "; "))

	// 更新执行记录
	config.DB.Model(&execution).Updates(map[string]interface{}{
		"nomad_job_id": execution.NomadJobID,
		"status":       execution.Status,
		"result":       execution.Result,
	})

	// 更新任务最后执行时间
	now := time.Now()
	customTask.LastRun = &now
	config.DB.Save(&customTask)
}

// generateNomadJobSpec 生成Nomad Job配置
func (h *CustomTaskHandler) generateNomadJobSpec(customTask *task.CustomTask, hostName string) (string, error) {
	// 解析参数
	var params map[string]interface{}
	if customTask.Parameters != "" {
		if err := json.Unmarshal([]byte(customTask.Parameters), &params); err != nil {
			return "", fmt.Errorf("failed to parse parameters: %v", err)
		}
	}

	// 渲染脚本模板
	renderedScript, err := h.renderScriptTemplate(customTask.Script, params)
	if err != nil {
		return "", fmt.Errorf("failed to render script template: %v", err)
	}

	// 根据脚本类型生成不同的Job配置
	switch customTask.ScriptType {
	case "bash", "shell":
		return h.generateShellJob(customTask, renderedScript, hostName)
	case "python":
		return h.generatePythonJob(customTask, renderedScript, hostName)
	default:
		return "", fmt.Errorf("unsupported script type: %s", customTask.ScriptType)
	}
}

// renderScriptTemplate 渲染脚本模板
func (h *CustomTaskHandler) renderScriptTemplate(scriptTemplate string, params map[string]interface{}) (string, error) {
	tmpl, err := template.New("script").Parse(scriptTemplate)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateShellJob 生成Shell脚本Job
func (h *CustomTaskHandler) generateShellJob(customTask *task.CustomTask, script string, hostName string) (string, error) {
	// 生成唯一的job名称，使用主机名前缀
	jobName := fmt.Sprintf("shell-task-%s-%s", hostName, time.Now().Format("20060102150405"))

	// 生成节点约束（针对单个主机）
	nodeConstraints := h.generateNodeConstraints([]string{hostName})

	// 生成Job配置
	jobSpec := fmt.Sprintf(`
job "%s" {
  datacenters = ["dc1"]
  type = "batch"
  
  %s
  
  group "script" {
    count = 1
    
    task "script" {
      driver = "raw_exec"
      
      config {
        command = "/bin/bash"
        args = ["local/script.sh"]
      }
      
      resources {
        cpu    = 200
        memory = 256
      }
      
      restart {
        attempts = 0
        mode = "fail"
      }
      
      template {
        data = <<EOH
#!/bin/bash
set -e

echo "Starting custom task: %s"
echo "Task ID: %d"
echo "Target host: %s"
echo "Execution time: $(date)"

# 设置超时
timeout %d bash << 'SCRIPT_EOF'
%s
SCRIPT_EOF

EXIT_CODE=$?
echo "Task completed with exit code: $EXIT_CODE"
exit $EXIT_CODE
EOH
        destination = "local/script.sh"
        perms        = "0755"
      }
    }
  }
}`,
		jobName,
		nodeConstraints,
		customTask.Name,
		customTask.ID,
		hostName,
		customTask.Timeout,
		script,
	)

	return jobSpec, nil
}

// generatePythonJob 生成Python脚本Job
func (h *CustomTaskHandler) generatePythonJob(customTask *task.CustomTask, script string, hostName string) (string, error) {
	// 生成唯一的job名称，使用主机名前缀
	jobName := fmt.Sprintf("python-task-%s-%s", hostName, time.Now().Format("20060102150405"))

	// 生成节点约束（针对单个主机）
	nodeConstraints := h.generateNodeConstraints([]string{hostName})

	// 生成Job配置
	jobSpec := fmt.Sprintf(`
job "%s" {
  datacenters = ["dc1"]
  type = "batch"
  
  %s
  
  group "script" {
    count = 1
    
    task "script" {
      driver = "raw_exec"
      
      config {
        command = "/usr/bin/python3"
        args = ["local/script.py"]
      }
      
      resources {
        cpu    = 200
        memory = 256
      }
      
      restart {
        attempts = 0
        mode = "fail"
      }
      
      template {
        data = <<EOH
#!/usr/bin/env python3
import sys
import time
import signal
import os

def timeout_handler(signum, frame):
    print("Task timeout after %d seconds")
    sys.exit(1)

# 设置超时
signal.signal(signal.SIGALRM, timeout_handler)
signal.alarm(%d)

print("Starting custom task: %s")
print("Task ID: %d")
print("Target host: %s")
print("Execution time:", time.strftime("%%Y-%%m-%%d %%H:%%M:%%S"))

try:
%s
    print("Task completed successfully")
except Exception as e:
    print(f"Task failed with error: {e}")
    sys.exit(1)
finally:
    signal.alarm(0)
EOH
        destination = "local/script.py"
        perms        = "0755"
      }
    }
  }
}`,
		jobName,
		nodeConstraints,
		customTask.Timeout,
		customTask.Timeout,
		customTask.Name,
		customTask.ID,
		hostName,
		script,
	)

	return jobSpec, nil
}

// isValidScriptType 验证脚本类型
func (h *CustomTaskHandler) isValidScriptType(scriptType string) bool {
	validTypes := []string{"bash", "shell", "python"}
	for _, t := range validTypes {
		if t == scriptType {
			return true
		}
	}
	return false
}

// Handler_ListCustomTasks 列出自定义任务
func (h *CustomTaskHandler) Handler_ListCustomTasks(c fiber.Ctx) error {
	// 获取分页参数
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("perPage"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页面大小
	}

	// 获取搜索参数
	name := c.Query("name")
	scriptType := c.Query("script_type")
	status := c.Query("status")

	// 构建查询
	query := config.DB.Model(&task.CustomTask{})

	// 应用搜索条件
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if scriptType != "" {
		query = query.Where("script_type = ?", scriptType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to count tasks", err.Error(), nil)
	}

	// 分页查询
	var tasks []task.CustomTask
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list tasks", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"items":   tasks,
		"total":   total,
		"page":    page,
		"perPage": pageSize,
	})
}

// Handler_GetCustomTask 获取自定义任务详情
func (h *CustomTaskHandler) Handler_GetCustomTask(c fiber.Ctx) error {
	taskID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid task id", err.Error(), nil)
	}

	var customTask task.CustomTask
	result, err := h.ListByID(uint(taskID))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusNotFound, 1, "task not found", err.Error(), nil)
	}
	customTask = *result

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", customTask)
}

// Handler_DeleteCustomTask 删除自定义任务
func (h *CustomTaskHandler) Handler_DeleteCustomTask(c fiber.Ctx) error {
	taskID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid task id", err.Error(), nil)
	}

	if err := h.Delete(uint(taskID)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete custom task", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "custom task deleted successfully", "", nil)
}

// Handler_UpdateCustomTask 更新自定义任务
func (h *CustomTaskHandler) Handler_UpdateCustomTask(c fiber.Ctx) error {
	taskID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid task id", err.Error(), nil)
	}

	var payload task.CustomTaskPayload
	if err := c.Bind().Body(&payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid request body", err.Error(), nil)
	}

	var customTask task.CustomTask
	result, err := h.ListByID(uint(taskID))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusNotFound, 1, "task not found", err.Error(), nil)
	}
	customTask = *result

	// 更新字段
	customTask.Name = payload.Name
	customTask.Description = payload.Description
	customTask.ScriptType = payload.ScriptType
	customTask.Script = payload.Script
	customTask.TargetHosts = payload.TargetHosts
	//customTask.Schedule = payload.Schedule
	customTask.Timeout = payload.Timeout
	customTask.RetryCount = payload.RetryCount

	customTask.Parameters = payload.Parameters // 直接使用字符串，因为已经是 JSON 字符串格式

	if err := h.Update(customTask.ID, &customTask); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to update custom task", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "custom task updated successfully", "", customTask)
}

// parseTargetHosts 解析目标主机列表
func (h *CustomTaskHandler) parseTargetHosts(targetHostsStr string) []string {
	if targetHostsStr == "" {
		return []string{}
	}

	hosts := strings.Split(targetHostsStr, ",")
	var result []string
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if host != "" {
			result = append(result, host)
		}
	}
	return result
}

// generateNodeConstraints 生成节点约束
func (h *CustomTaskHandler) generateNodeConstraints(targetHosts []string) string {
	if len(targetHosts) == 0 {
		return ""
	}
	patterns := append([]string{}, targetHosts...)
	// 使用regexp操作符进行节点约束
	pattern := strings.Join(patterns, "|")
	return fmt.Sprintf(`constraint {
  attribute = "${attr.unique.hostname}"
  operator  = "regexp"
  value     = "^(%s)$"
}`, pattern)
}

// formatHCL 格式化HCL配置
func (h *CustomTaskHandler) formatHCL(hclContent string) (string, error) {
	// 使用hclwrite解析并格式化HCL
	f, diags := hclwrite.ParseConfig([]byte(hclContent), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return "", fmt.Errorf("failed to parse HCL: %v", diags)
	}

	return string(f.Bytes()), nil
}

// saveJobHCLToConsulWithKey 保存Job HCL到Consul（指定key）
func (h *CustomTaskHandler) saveJobHCLToConsulWithKey(customTask *task.CustomTask, jobSpec string, key string) error {
	// 格式化HCL配置
	formattedJobSpec, err := h.formatHCL(jobSpec)
	if err != nil {
		slog.Warn("Failed to format HCL, using original", "error", err)
		formattedJobSpec = jobSpec
	}

	// 保存到Consul
	kvPair := &consulapi.KVPair{
		Key:   key,
		Value: []byte(formattedJobSpec),
	}

	_, err = config.ConsulCli.KV().Put(kvPair, nil)
	if err != nil {
		return fmt.Errorf("failed to save job HCL to Consul: %v", err)
	}

	slog.Info("Job HCL saved to Consul", "key", key, "task_id", customTask.ID)
	return nil
}

// Handler_GetExecutionStatus 获取执行状态
func (h *CustomTaskHandler) Handler_GetExecutionStatus(c fiber.Ctx) error {
	executionID, err := strconv.Atoi(c.Params("execution_id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid execution id", err.Error(), nil)
	}

	execution, err := h.Monitor.GetExecutionStatus(uint(executionID))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusNotFound, 1, "execution not found", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", execution)
}

// Handler_GetExecutionLogs 获取执行日志
func (h *CustomTaskHandler) Handler_GetExecutionLogs(c fiber.Ctx) error {
	executionID, err := strconv.Atoi(c.Params("execution_id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid execution id", err.Error(), nil)
	}

	logs, err := h.Monitor.GetExecutionLogs(uint(executionID))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to get logs", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"logs": logs,
	})
}

// Handler_StopExecution 停止执行
func (h *CustomTaskHandler) Handler_StopExecution(c fiber.Ctx) error {
	executionID, err := strconv.Atoi(c.Params("execution_id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid execution id", err.Error(), nil)
	}

	if err := h.Monitor.StopJob(uint(executionID)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to stop execution", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "execution stopped successfully", "", nil)
}

// Handler_ListExecutions 列出任务的执行记录
func (h *CustomTaskHandler) Handler_ListExecutions(c fiber.Ctx) error {
	taskID, err := strconv.Atoi(c.Params("task_id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid task id", err.Error(), nil)
	}

	var executions []task.CustomTaskExecution
	if err := config.DB.Where("task_id = ?", taskID).Order("created_at DESC").Find(&executions).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list executions", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"executions": executions,
	})
}

// recordFailedJob 记录失败的任务
func (h *CustomTaskHandler) recordFailedJob(mu *sync.Mutex, failCount *int, failedJobs *[]string, key string) {
	mu.Lock()
	defer mu.Unlock()
	*failCount++
	*failedJobs = append(*failedJobs, key)
}

// recordSuccessJob 记录成功的任务
func (h *CustomTaskHandler) recordSuccessJob(mu *sync.Mutex, successCount *int, successJobs *[]string, key string) {
	mu.Lock()
	defer mu.Unlock()
	*successCount++
	*successJobs = append(*successJobs, key)
}
