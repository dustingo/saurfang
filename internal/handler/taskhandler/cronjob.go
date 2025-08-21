package taskhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"saurfang/internal/config"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/models/task"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strings"
	"time"

	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/hibiken/asynq"
)

type CronjobHandler struct {
	base.BaseGormRepository[task.CronJobs]
}

// Handler_CreateCronjobTask 创建计划任务（支持多种任务类型）
func (j *CronjobHandler) Handler_CreateCronjobTask(c fiber.Ctx) error {
	var payload task.CronJobPayload
	if err := c.Bind().Body(&payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "创建计划任务失败", err.Error(), nil)
	}

	// 验证任务类型
	if !j.isValidTaskType(payload.TaskType) {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "创建计划任务失败", "invalid task type", nil)
	}

	// 根据任务类型进行验证
	if err := j.validateTaskPayload(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "创建计划任务失败", err.Error(), nil)
	}

	var cronJob task.CronJobs
	cronJob.TaskName = payload.TaskName
	cronJob.Spec = payload.Spec
	cronJob.TaskType = payload.TaskType

	// 根据任务类型设置相应字段
	switch payload.TaskType {
	case task.TaskTypeCustom:
		cronJob.CustomTaskID = payload.CustomTaskID
	case task.TaskTypeServer:
		cronJob.ServerIDs = strings.Join(payload.ServerIDs, ",")
		cronJob.ServerOperation = payload.ServerOperation
	}

	if err := j.Create(&cronJob); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "创建计划任务失败", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusCreated, 0, "创建计划任务成功", "", fiber.Map{
		"data": cronJob,
	})
}

// Handler_UpdateCronjobTask 更新计划任务
func (j *CronjobHandler) Handler_UpdateCronjobTask(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var payload task.CronJobPayload
	if err := c.Bind().Body(&payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "更新计划任务失败", err.Error(), nil)
	}

	// 验证任务类型
	if !j.isValidTaskType(payload.TaskType) {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "任务类型错误", "invalid task type", nil)
	}

	// 根据任务类型进行验证
	if err := j.validateTaskPayload(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "更新计划任务失败", err.Error(), nil)
	}

	var cronJob task.CronJobs
	result, err := j.ListByID(uint(id))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusNotFound, 1, "更新计划任务失败", "cron job not found", nil)
	}
	cronJob = *result

	// 更新字段
	cronJob.TaskName = payload.TaskName
	cronJob.Spec = payload.Spec
	cronJob.TaskType = payload.TaskType
	cronJob.TaskStatus = payload.TaskStatus
	// 根据任务类型设置相应字段
	switch payload.TaskType {
	case task.TaskTypeCustom:
		cronJob.CustomTaskID = payload.CustomTaskID
		cronJob.ServerIDs = ""
		cronJob.ServerOperation = ""
	case task.TaskTypeServer:
		cronJob.CustomTaskID = nil
		cronJob.ServerIDs = strings.Join(payload.ServerIDs, ",")
		cronJob.ServerOperation = payload.ServerOperation
	}

	// 使用 Select 明确指定要更新的字段，包括零值字段 TaskStatus
	if err := j.DB.Model(&cronJob).Where("id = ?", uint(id)).Select("task_name", "spec", "task_type", "task_status", "custom_task_id", "server_ids", "server_operation").Updates(&cronJob).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "更新计划任务失败", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "更新计划任务成功", "", fiber.Map{
		"data": cronJob,
	})
}

// Handler_DeleteCronjobTask 删除计划任务
func (j *CronjobHandler) Handler_DeleteCronjobTask(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := j.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "删除计划任务失败", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "删除计划任务成功", "", nil)
}

// Handler_ShowCronjobTask 获取计划任务列表（支持搜索）
func (j *CronjobHandler) Handler_ShowCronjobTask(c fiber.Ctx) error {
	// 获取查询参数
	taskName := c.Query("task_name")
	taskType := c.Query("task_type")
	taskStatus := c.Query("task_status")

	// 构建查询
	query := config.DB.Preload("CustomTask")

	// 添加搜索条件
	if taskName != "" {
		query = query.Where("task_name LIKE ?", "%"+taskName+"%")
	}
	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}
	if taskStatus != "" {
		query = query.Where("task_status = ?", taskStatus)
	}

	// 执行查询
	var cronjobs []task.CronJobs
	if err := query.Find(&cronjobs).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "获取计划任务列表失败", err.Error(), nil)
	}

	// 格式化数据，为amis CRUD提供更好的展示
	var formattedJobs []map[string]interface{}
	for _, job := range cronjobs {
		formattedJob := map[string]interface{}{
			"id":          job.ID,
			"task_name":   job.TaskName,
			"spec":        job.Spec,
			"task_type":   job.TaskType,
			"task_status": job.TaskStatus,
			"created_at":  job.CreatedAt,
			"updated_at":  job.UpdatedAt,
		}

		// 根据任务类型添加特定字段
		switch job.TaskType {
		case task.TaskTypeCustom:
			if job.CustomTask != nil {
				formattedJob["custom_task"] = map[string]interface{}{
					"id":          job.CustomTask.ID,
					"name":        job.CustomTask.Name,
					"script_type": job.CustomTask.ScriptType,
					"description": job.CustomTask.Description,
				}
			}
		case task.TaskTypeServer:
			// 解析服务器ID列表
			serverIDs := []string{}
			if job.ServerIDs != "" {
				serverIDs = strings.Split(job.ServerIDs, ",")
			}

			formattedJob["server_ids"] = serverIDs
			formattedJob["server_operation"] = job.ServerOperation
		}

		// 添加最后执行时间
		if job.LastExecution != nil {
			formattedJob["last_execution"] = job.LastExecution
		}

		formattedJobs = append(formattedJobs, formattedJob)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "获取计划任务列表成功", "", fiber.Map{
		"data": formattedJobs,
	})
}

// Handler_ResetCronjobStatus 重置计划任务状态
func (j *CronjobHandler) Handler_ResetCronjobStatus(c fiber.Ctx) error {
	id := c.Params("id")
	if err := j.DB.Model(&task.CronJobs{}).Where("id = ?", id).Update("task_status", 0).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "重置计划任务状态失败", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "重置计划任务状态成功", "", nil)
}

// Handler_GetAvailableCustomTasks 获取可用的自定义任务列表
func (j *CronjobHandler) Handler_GetAvailableCustomTasks(c fiber.Ctx) error {
	var customTasks []task.CustomTask
	if err := config.DB.Where("status = ?", "active").Find(&customTasks).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "获取可用的自定义任务列表失败", err.Error(), nil)
	}

	// 只返回基本信息
	var result []map[string]interface{}
	for _, ct := range customTasks {
		result = append(result, map[string]interface{}{
			"id":          ct.ID,
			"name":        ct.Name,
			"description": ct.Description,
			"script_type": ct.ScriptType,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    result,
	})
}

// Handler_GetAvailableServers 获取可用的游戏服务器列表
func (j *CronjobHandler) Handler_GetAvailableServers(c fiber.Ctx) error {
	var games []gameserver.Games
	if err := config.DB.Preload("Channel").Find(&games).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "获取可用的游戏服务器列表失败", err.Error(), nil)
	}

	var result []map[string]interface{}
	for _, game := range games {
		channelName := ""
		if game.Channel != nil {
			channelName = game.Channel.Name
		}

		result = append(result, map[string]interface{}{
			"id":           game.ID,
			"name":         game.Name,
			"server_id":    game.ServerID,
			"status":       game.Status,
			"channel_name": channelName,
		})
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "获取可用的游戏服务器列表成功", "", fiber.Map{
		"data": result,
	})
}

// isValidTaskType 验证任务类型是否有效
func (j *CronjobHandler) isValidTaskType(taskType string) bool {
	validTypes := []string{task.TaskTypeCustom, task.TaskTypeServer}
	for _, validType := range validTypes {
		if taskType == validType {
			return true
		}
	}
	return false
}

// validateTaskPayload 根据任务类型验证请求参数
func (j *CronjobHandler) validateTaskPayload(payload task.CronJobPayload) error {
	switch payload.TaskType {
	case task.TaskTypeCustom:
		if payload.CustomTaskID == nil || *payload.CustomTaskID == 0 {
			return fmt.Errorf("custom_task_id is required for custom task type")
		}
		// 验证 CustomTask 是否存在
		var customTask task.CustomTask
		if err := config.DB.First(&customTask, *payload.CustomTaskID).Error; err != nil {
			return fmt.Errorf("custom task not found: %v", err)
		}
	case task.TaskTypeServer:
		if len(payload.ServerIDs) == 0 {
			return fmt.Errorf("server_ids is required for server operation task type")
		}
		if payload.ServerOperation == "" {
			return fmt.Errorf("server_operation is required for server operation task type")
		}
		// 验证服务器操作类型
		validOps := []string{task.ServerOpStart, task.ServerOpStop, task.ServerOpRestart}
		isValidOp := false
		for _, op := range validOps {
			if payload.ServerOperation == op {
				isValidOp = true
				break
			}
		}
		if !isValidOp {
			return fmt.Errorf("invalid server operation: %s", payload.ServerOperation)
		}
		// 验证服务器ID是否存在
		for _, serverID := range payload.ServerIDs {
			var game gameserver.Games
			if err := config.DB.Where("server_id = ?", serverID).First(&game).Error; err != nil {
				return fmt.Errorf("server not found: %s", serverID)
			}
		}
	}
	return nil
}

// CustomTaskHandler 处理自定义任务
func CustonCronjobHandler(ctx context.Context, at *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(at.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	customTaskID, ok := payload["custom_task_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid custom_task_id in payload")
	}
	// 获取 CustomTask
	var customTask task.CustomTask
	if err := config.DB.First(&customTask, uint(customTaskID)).Error; err != nil {
		return fmt.Errorf("custom task not found: %v", err)
	}
	// 创建执行记录
	execution := task.CustomTaskExecution{
		TaskID:        uint(customTaskID),
		Status:        "pending",
		StartTime:     time.Now(),
		CheckInterval: 10, // 10秒检查一次
		MaxCheckCount: 60, // 最多检查60次（10分钟）
	}
	if err := config.DB.Create(&execution).Error; err != nil {
		return fmt.Errorf("failed to create execution record: %v", err)
	}
	customTaskExecuter := NewCustomTaskHandler()
	go customTaskExecuter.ExecuteCustomTaskAsync(&customTask, &execution)
	costomTaskMonitor, err := pkg.NewNomadMonitor()
	if err != nil {
		slog.Error("failed to create nomad job monitor", "error", err)
		return nil
	} else {
		go costomTaskMonitor.StartMonitoring(&execution)
	}
	return nil
}
