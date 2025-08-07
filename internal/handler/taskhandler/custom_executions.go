package taskhandler

import (
	"fmt"
	"saurfang/internal/config"
	"saurfang/internal/models/task"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type CustomTaskExecutionsHandler struct {
	base.BaseGormRepository[task.CustomTaskExecution]
}

func NewCustomTaskExecutionsHandler() *CustomTaskExecutionsHandler {
	// nomadClient, err := config.NewNomadClient()
	// if err != nil {
	// 	slog.Error("failed to create nomad client", "error", err)
	// 	os.Exit(1)
	// }
	return &CustomTaskExecutionsHandler{
		BaseGormRepository: base.BaseGormRepository[task.CustomTaskExecution]{DB: config.DB},
	}
}

// Handler_ListCustomExecutions 列出自定义任务执行记录(支持分页以及搜索)
func (c *CustomTaskExecutionsHandler) Handler_ListCustomExecutions(ctx fiber.Ctx) error {
	// 获取分页参数
	page, err := strconv.Atoi(ctx.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(ctx.Query("perPage"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页面大小
	}

	// 获取搜索参数
	status := ctx.Query("status")
	taskName := ctx.Query("task_name")
	startTimeFrom := ctx.Query("start_time_from")
	startTimeTo := ctx.Query("start_time_to")

	// 构建基础查询 - 只查询 custom_task_executions 表
	query := config.DB.Table("custom_task_executions")

	// 应用搜索条件 - 只针对 custom_task_executions 表的字段
	var conditions []string
	var args []any

	if status != "" {
		conditions = append(conditions, "custom_task_executions.status = ?")
		args = append(args, status)
	}

	if startTimeFrom != "" {
		conditions = append(conditions, "custom_task_executions.start_time >= ?")
		args = append(args, startTimeFrom)
	}

	if startTimeTo != "" {
		conditions = append(conditions, "custom_task_executions.start_time <= ?")
		args = append(args, startTimeTo)
	}

	// 如果有任务名称搜索，需要先查询任务ID
	if taskName != "" {
		// 查询匹配任务名称的任务ID列表
		var taskIDs []uint
		taskQuery := config.DB.Table("custom_tasks").
			Select("id").
			Where("name LIKE ?", "%"+taskName+"%")

		if err := taskQuery.Find(&taskIDs).Error; err != nil {
			return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "failed to query task IDs", err.Error(), nil)
		}

		if len(taskIDs) > 0 {
			conditions = append(conditions, "custom_task_executions.task_id IN ?")
			args = append(args, taskIDs)
		} else {
			// 如果没有找到匹配的任务，直接返回空结果
			return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", fiber.Map{
				"items":   []any{},
				"total":   0,
				"page":    page,
				"perPage": pageSize,
			})
		}
	}

	// 应用搜索条件
	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " AND "), args...)
	}

	// 按创建时间倒序排列
	query = query.Order("custom_task_executions.created_at DESC")

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "failed to count executions", err.Error(), nil)
	}

	// 分页查询
	var executions []map[string]interface{}
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&executions).Error; err != nil {
		return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "failed to list executions", err.Error(), nil)
	}

	// 格式化返回数据
	var result []map[string]interface{}

	// 收集所有任务ID，用于批量查询任务名称
	var taskIDs []uint
	for _, execution := range executions {
		if taskID, ok := execution["task_id"].(uint); ok {
			taskIDs = append(taskIDs, taskID)
		}
	}

	// 批量查询任务名称
	taskNameMap := make(map[uint]string)
	if len(taskIDs) > 0 {
		var tasks []map[string]interface{}
		taskQuery := config.DB.Table("custom_tasks").
			Select("id, name").
			Where("id IN ?", taskIDs)

		if err := taskQuery.Find(&tasks).Error; err != nil {
			return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "failed to query task names", err.Error(), nil)
		}

		for _, task := range tasks {
			if taskID, ok := task["id"].(uint); ok {
				if taskName, ok := task["name"].(string); ok {
					taskNameMap[taskID] = taskName
				}
			}
		}
	}

	for _, execution := range executions {
		// 计算执行时长
		var duration string
		if execution["start_time"] != nil && execution["end_time"] != nil {
			startTime, ok1 := execution["start_time"].(time.Time)
			endTime, ok2 := execution["end_time"].(time.Time)
			if ok1 && ok2 {
				durationSeconds := endTime.Sub(startTime).Seconds()
				duration = fmt.Sprintf("%.1f秒", durationSeconds)
			} else {
				duration = "-"
			}
		} else {
			duration = "-"
		}

		// 获取任务名称
		var taskName string
		if taskID, ok := execution["task_id"].(uint); ok {
			if name, exists := taskNameMap[taskID]; exists {
				taskName = name
			} else {
				taskName = "未知任务"
			}
		}

		result = append(result, map[string]interface{}{
			"id":           execution["id"],
			"task_id":      execution["task_id"],
			"task_name":    taskName,
			"status":       execution["status"],
			"start_time":   execution["start_time"],
			"end_time":     execution["end_time"],
			"duration":     duration,
			"exit_code":    execution["exit_code"],
			"nomad_job_id": execution["nomad_job_id"],
			"result":       execution["result"],
			"err_msg":      execution["error_msg"], // 匹配AMIS配置中的字段名
			"created_at":   execution["created_at"],
		})
	}

	return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", fiber.Map{
		"rows":    result,
		"count":   total,
		"page":    page,
		"perPage": pageSize,
	})
}

// Handler_DeleteCustomExecutions 删除自定义任务执行记录
func (c *CustomTaskExecutionsHandler) Handler_DeleteCustomExecutions(ctx fiber.Ctx) error {
	id, _ := strconv.Atoi(ctx.Params("id"))
	err := c.Delete(uint(id))
	if err != nil {
		return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "failed to delete custom execution", err.Error(), nil)
	}
	return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", fiber.Map{
		"data": []any{},
	})
}
