package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"saurfang/internal/config"
	"saurfang/internal/models/task"
	"strings"
	"time"

	nomadapi "github.com/hashicorp/nomad/api"
)

// NomadMonitor Nomad 任务监控服务
type NomadMonitor struct {
	NomadClient *nomadapi.Client
}

// NewNomadMonitor 创建新的 Nomad 监控服务
func NewNomadMonitor() (*NomadMonitor, error) {
	// nomadClient, err := config.NewNomadClient()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create nomad client: %v", err)
	// }
	return &NomadMonitor{
		NomadClient: config.NomadCli,
	}, nil
}

// StartMonitoring 开始监控执行记录
func (m *NomadMonitor) StartMonitoring(execution *task.CustomTaskExecution) {
	go m.monitorExecution(execution)
}

// monitorExecution 监控单个执行记录
func (m *NomadMonitor) monitorExecution(execution *task.CustomTaskExecution) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(execution.MaxCheckCount*execution.CheckInterval)*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Duration(execution.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// 超时，标记为失败
			m.updateExecutionStatus(execution, "failed", "Monitoring timeout", -1)
			return
		case <-ticker.C:
			// 检查执行状态
			if m.checkExecutionStatus(execution) {
				return // 任务完成或失败，停止监控
			}
		}
	}
}

// checkExecutionStatus 检查执行状态
func (m *NomadMonitor) checkExecutionStatus(execution *task.CustomTaskExecution) bool {
	execution.CheckCount++
	now := time.Now()
	execution.LastCheckTime = &now

	// 更新检查次数
	config.DB.Model(execution).Update("check_count", execution.CheckCount)
	config.DB.Model(execution).Update("last_check_time", execution.LastCheckTime)

	// 检查是否超过最大检查次数
	if execution.CheckCount >= execution.MaxCheckCount {
		m.updateExecutionStatus(execution, "failed", "Exceeded maximum check count", -1)
		return true
	}

	// 解析多个Job ID（用逗号分隔）
	jobIDs := strings.Split(execution.NomadJobID, ",")
	if len(jobIDs) == 0 {
		m.updateExecutionStatus(execution, "failed", "No job IDs found", -1)
		return true
	}

	// 检查所有Job的状态
	var allDead bool = true
	var allComplete bool = true
	var allFailed bool = true
	var anyRunning bool = false
	var anyPending bool = false
	var jobStatuses []string
	var failedJobs []string

	for _, jobID := range jobIDs {
		jobID = strings.TrimSpace(jobID)
		if jobID == "" {
			continue
		}

		// 获取 Job 信息
		job, _, err := m.NomadClient.Jobs().Info(jobID, &nomadapi.QueryOptions{})

		if err != nil {
			slog.Error("Failed to get job info", "job_id", jobID, "error", err)
			// 如果无法获取job信息，假设还在运行
			anyRunning = true
			allComplete = false
			allFailed = false
			allDead = false
			continue
		}

		jobStatus := *job.Status
		jobStatuses = append(jobStatuses, fmt.Sprintf("%s:%s", jobID, jobStatus))

		// 检查 Job 状态
		switch jobStatus {
		case "pending":
			anyPending = true
			allComplete = false
			allFailed = false
			allDead = false

		case "running":
			anyRunning = true
			allComplete = false
			allFailed = false
			allDead = false

		case "complete", "dead":
			allFailed = false
			// 检查是否真的完成
			if m.handleJobComplete(execution, job) {
				// 任务成功完成
			} else {
				allComplete = false
			}

		case "failed", "canceled":
			failedJobs = append(failedJobs, jobID)
			allComplete = false
			// 检查是否所有job都失败了
			if len(failedJobs) == len(jobIDs) {
				allFailed = true
			}

		default:
			// 未知状态，假设还在运行
			anyRunning = true
			allComplete = false
			allFailed = false
			allDead = false
		}
	}

	// 更新 Nomad Job 状态（记录所有job的状态）
	execution.NomadJobStatus = strings.Join(jobStatuses, "; ")
	config.DB.Model(execution).Update("nomad_job_status", execution.NomadJobStatus)

	// 根据所有Job的状态决定整体执行状态
	if allComplete {
		execution.Status = "completed"
		config.DB.Model(execution).Update("status", execution.Status)
		return true
	}

	if allFailed {
		errorMsg := fmt.Sprintf("All jobs failed: %s", strings.Join(failedJobs, ", "))
		m.updateExecutionStatus(execution, "failed", errorMsg, -1)
		return true
	}

	if allDead {
		execution.Status = "dead"
		config.DB.Model(execution).Update("status", execution.Status)
		return true
	}

	if anyRunning {
		execution.Status = "running"
		config.DB.Model(execution).Update("status", execution.Status)
		return false
	}

	if anyPending {
		execution.Status = "pending"
		config.DB.Model(execution).Update("status", execution.Status)
		return false
	}

	// 其他情况，继续监控
	return false
}

// handleJobComplete 处理 Job 完成
func (m *NomadMonitor) handleJobComplete(execution *task.CustomTaskExecution, job *nomadapi.Job) bool {
	// 获取 Allocation 信息 - 使用当前检查的job ID
	allocs, _, err := m.NomadClient.Jobs().Allocations(*job.ID, false, &nomadapi.QueryOptions{})
	if err != nil {
		slog.Error("Failed to get job allocations", "job_id", *job.ID, "error", err)
		// 不立即失败，继续监控其他job
		return false
	}

	if len(allocs) == 0 {
		slog.Warn("No allocations found for job", "job_id", *job.ID)
		return false
	}

	// 获取第一个 Allocation 的详细信息
	alloc, _, err := m.NomadClient.Allocations().Info(allocs[0].ID, &nomadapi.QueryOptions{})
	if err != nil {
		slog.Error("Failed to get allocation info", "alloc_id", allocs[0].ID, "error", err)
		return false
	}

	// 更新 Allocation 相关信息（记录第一个完成的job的信息）
	if execution.NomadAllocID == "" {
		execution.NomadAllocID = alloc.ID
		execution.NomadNodeID = alloc.NodeID
		config.DB.Model(execution).Updates(map[string]interface{}{
			"nomad_alloc_id": execution.NomadAllocID,
			"nomad_node_id":  execution.NomadNodeID,
		})
	}

	// 检查 Allocation 状态
	switch alloc.ClientStatus {
	case "complete":
		return m.handleAllocationComplete(execution, alloc)
	case "failed":
		// 单个job失败，但不立即结束整个执行
		slog.Warn("Job allocation failed", "job_id", *job.ID, "alloc_id", alloc.ID)
		return false
	default:
		slog.Warn("Unexpected allocation status", "alloc_id", alloc.ID, "status", alloc.ClientStatus)
		return false
	}
}

// handleAllocationComplete 处理 Allocation 完成
func (m *NomadMonitor) handleAllocationComplete(execution *task.CustomTaskExecution, alloc *nomadapi.Allocation) bool {
	// 获取任务日志
	logs, err := m.getTaskLogs(alloc.ID, "script")
	if err != nil {
		slog.Error("Failed to get task logs", "alloc_id", alloc.ID, "error", err)
		// 即使获取日志失败，也继续处理
	}

	// 获取任务退出码
	exitCode := -1
	if len(alloc.TaskStates) > 0 {
		if _, exists := alloc.TaskStates["script"]; exists {
			// 如果任务状态存在，认为任务已完成
			// 退出码默认为 0（成功），实际退出码可以从日志中解析
			exitCode = 0
		}
	}

	// 更新执行状态
	if exitCode == 0 {
		m.updateExecutionStatus(execution, "success", logs, exitCode)
	} else {
		m.updateExecutionStatus(execution, "failed", logs, exitCode)
	}

	return true
}

// getTaskLogs 获取任务日志
func (m *NomadMonitor) getTaskLogs(allocID, taskName string) (string, error) {
	var logs []string

	// 首先获取 Allocation 信息
	alloc, _, err := m.NomadClient.Allocations().Info(allocID, &nomadapi.QueryOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get allocation info: %v", err)
	}

	// 获取标准输出日志
	stdoutChan, errChan := m.NomadClient.AllocFS().Logs(alloc, false, taskName, "stdout", "start", 0, nil, &nomadapi.QueryOptions{})

	// 读取标准输出日志
	for frame := range stdoutChan {
		if frame.Data != nil {
			logs = append(logs, string(frame.Data))
		}
	}

	// 检查是否有错误
	select {
	case err := <-errChan:
		if err != nil {
			return "", fmt.Errorf("failed to get stdout logs: %v", err)
		}
	default:
	}

	// 获取错误日志
	stderrChan, stderrErrChan := m.NomadClient.AllocFS().Logs(alloc, false, taskName, "stderr", "start", 0, nil, &nomadapi.QueryOptions{})

	logs = append(logs, "\n=== STDERR ===\n")
	// 读取错误日志
	for frame := range stderrChan {
		if frame.Data != nil {
			logs = append(logs, string(frame.Data))
		}
	}

	// 检查是否有错误
	select {
	case err := <-stderrErrChan:
		if err != nil {
			// 即使获取 stderr 失败，也继续处理
			slog.Error("Failed to get stderr logs", "error", err)
		}
	default:
	}

	return strings.Join(logs, ""), nil
}

// updateExecutionStatus 更新执行状态
func (m *NomadMonitor) updateExecutionStatus(execution *task.CustomTaskExecution, status, result string, exitCode int) {
	now := time.Now()
	execution.Status = status
	execution.Result = result
	execution.ExitCode = exitCode
	execution.EndTime = &now

	// 更新数据库
	config.DB.Model(execution).Updates(map[string]interface{}{
		"status":    execution.Status,
		"result":    execution.Result,
		"exit_code": execution.ExitCode,
		"end_time":  execution.EndTime,
	})

	slog.Info("Execution status updated",
		"execution_id", execution.ID,
		"status", status,
		"exit_code", exitCode)
}

// GetExecutionStatus 获取执行状态
func (m *NomadMonitor) GetExecutionStatus(executionID uint) (*task.CustomTaskExecution, error) {
	var execution task.CustomTaskExecution
	if err := config.DB.First(&execution, executionID).Error; err != nil {
		return nil, err
	}
	return &execution, nil
}

// GetExecutionLogs 获取执行日志
func (m *NomadMonitor) GetExecutionLogs(executionID uint) (string, error) {
	var execution task.CustomTaskExecution
	if err := config.DB.First(&execution, executionID).Error; err != nil {
		return "", err
	}

	if execution.NomadJobID == "" {
		return "", fmt.Errorf("no job IDs found for execution %d", executionID)
	}

	// 解析多个Job ID（用逗号分隔）
	jobIDs := strings.Split(execution.NomadJobID, ",")
	var allLogs []string

	for _, jobID := range jobIDs {
		jobID = strings.TrimSpace(jobID)
		if jobID == "" {
			continue
		}

		// 获取Job的Allocation信息
		allocs, _, err := m.NomadClient.Jobs().Allocations(jobID, false, &nomadapi.QueryOptions{})
		if err != nil {
			slog.Error("Failed to get job allocations for logs", "job_id", jobID, "error", err)
			allLogs = append(allLogs, fmt.Sprintf("=== Job %s ===\nFailed to get allocations: %v\n", jobID, err))
			continue
		}

		if len(allocs) == 0 {
			allLogs = append(allLogs, fmt.Sprintf("=== Job %s ===\nNo allocations found\n", jobID))
			continue
		}

		// 获取第一个Allocation的日志
		allocLogs, err := m.getTaskLogs(allocs[0].ID, "script")
		if err != nil {
			slog.Error("Failed to get task logs", "alloc_id", allocs[0].ID, "error", err)
			allLogs = append(allLogs, fmt.Sprintf("=== Job %s ===\nFailed to get logs: %v\n", jobID, err))
		} else {
			allLogs = append(allLogs, fmt.Sprintf("=== Job %s ===\n%s\n", jobID, allocLogs))
		}
	}

	return strings.Join(allLogs, "\n"), nil
}

// StopJob 停止 Job
func (m *NomadMonitor) StopJob(executionID uint) error {
	var execution task.CustomTaskExecution
	if err := config.DB.First(&execution, executionID).Error; err != nil {
		return err
	}

	if execution.NomadJobID == "" {
		return fmt.Errorf("no job ID found for execution %d", executionID)
	}

	// 解析多个Job ID（用逗号分隔）
	jobIDs := strings.Split(execution.NomadJobID, ",")
	var stoppedJobs []string
	var failedJobs []string

	for _, jobID := range jobIDs {
		jobID = strings.TrimSpace(jobID)
		if jobID == "" {
			continue
		}

		// 停止 Job
		_, _, err := m.NomadClient.Jobs().Deregister(jobID, false, &nomadapi.WriteOptions{})
		if err != nil {
			slog.Error("Failed to stop job", "job_id", jobID, "error", err)
			failedJobs = append(failedJobs, jobID)
		} else {
			stoppedJobs = append(stoppedJobs, jobID)
		}
	}

	// 更新执行状态
	var result string
	if len(failedJobs) == 0 {
		result = fmt.Sprintf("All jobs stopped successfully: %s", strings.Join(stoppedJobs, ", "))
	} else if len(stoppedJobs) == 0 {
		result = fmt.Sprintf("Failed to stop any jobs: %s", strings.Join(failedJobs, ", "))
	} else {
		result = fmt.Sprintf("Partially stopped: success=%s, failed=%s",
			strings.Join(stoppedJobs, ", "), strings.Join(failedJobs, ", "))
	}

	m.updateExecutionStatus(&execution, "canceled", result, -1)

	return nil
}
