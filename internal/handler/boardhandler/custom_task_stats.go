package boardhandler

import (
	"saurfang/internal/config"
	"saurfang/internal/models/task"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

// CustomTaskDetailedStats 自定义任务详细统计
type CustomTaskDetailedStats struct {
	TotalTasks           int64   `json:"total_tasks"`
	ActiveTasks          int64   `json:"active_tasks"`
	TotalExecutions      int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions     int64   `json:"failed_executions"`
	RunningExecutions    int64   `json:"running_executions"`
	AvgExecutionTime     float64 `json:"avg_execution_time"`
	SuccessRate          float64 `json:"success_rate"`
}

// CustomTaskExecutionHistory 自定义任务执行历史
type CustomTaskExecutionHistory struct {
	TaskID    uint       `json:"task_id"`
	TaskName  string     `json:"task_name"`
	Status    string     `json:"status"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Duration  float64    `json:"duration"`
	ErrorMsg  string     `json:"error_msg,omitempty"`
}

// Handler_CustomTaskDetailedStats 获取自定义任务详细统计
func Handler_CustomTaskDetailedStats(c fiber.Ctx) error {
	// 获取图表类型参数
	chartType := c.Query("chartType", "overview")

	var chartData interface{}

	switch chartType {
	case "overview":
		// 概览统计 - 返回原始统计数据
		stats := getCustomTaskDetailedStats()
		chartData = stats
	case "stats_pie":
		// 统计饼图
		stats := getCustomTaskDetailedStats()
		chartData = map[string]interface{}{
			"series": []map[string]interface{}{
				{
					"name":   "任务统计",
					"type":   "pie",
					"radius": "50%",
					"data": []map[string]interface{}{
						{"name": "总任务数", "value": stats.TotalTasks},
						{"name": "活跃任务", "value": stats.ActiveTasks},
						{"name": "总执行次数", "value": stats.TotalExecutions},
						{"name": "成功执行", "value": stats.SuccessfulExecutions},
						{"name": "失败执行", "value": stats.FailedExecutions},
						{"name": "运行中", "value": stats.RunningExecutions},
					},
					"label": map[string]interface{}{
						"show":      true,
						"formatter": "{b}: {c}({d}%)",
					},
				},
			},
		}
	case "execution_status_bar":
		// 执行状态柱状图
		stats := getCustomTaskDetailedStats()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": []string{"成功执行", "失败执行", "运行中"},
			},
			"yAxis": map[string]interface{}{
				"type": "value",
			},
			"series": []map[string]interface{}{
				{
					"name": "执行次数",
					"type": "bar",
					"data": []int64{stats.SuccessfulExecutions, stats.FailedExecutions, stats.RunningExecutions},
					"itemStyle": map[string]interface{}{
						"color": []string{"#67C23A", "#F56C6C", "#E6A23C"},
					},
				},
			},
		}
	case "performance_metrics":
		// 性能指标图 - 仪表盘
		stats := getCustomTaskDetailedStats()
		chartData = map[string]interface{}{
			"series": []map[string]interface{}{
				{
					"name":   "成功率",
					"type":   "gauge",
					"radius": "80%",
					"data": []map[string]interface{}{
						{
							"value": stats.SuccessRate,
							"name":  "成功率",
						},
					},
					"detail": map[string]interface{}{
						"formatter": "{value}%",
						"fontSize":  20,
						"color":     "#67C23A",
					},
					"title": map[string]interface{}{
						"fontSize": 14,
						"color":    "#666",
					},
					"axisLine": map[string]interface{}{
						"lineStyle": map[string]interface{}{
							"width": 20,
							"color": [][]interface{}{
								{0.8, "#67C23A"},
								{0.9, "#E6A23C"},
								{1.0, "#F56C6C"},
							},
						},
					},
					"pointer": map[string]interface{}{
						"itemStyle": map[string]interface{}{
							"color": "#67C23A",
						},
					},
					"max": 100,
				},
			},
		}
	case "performance_metrics_bar":
		// 性能指标柱状图
		stats := getCustomTaskDetailedStats()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": []string{"平均执行时间(秒)", "成功率(%)"},
			},
			"yAxis": map[string]interface{}{
				"type": "value",
			},
			"series": []map[string]interface{}{
				{
					"name": "指标值",
					"type": "bar",
					"data": []float64{stats.AvgExecutionTime, stats.SuccessRate},
					"itemStyle": map[string]interface{}{
						"color": []string{"#409EFF", "#67C23A"},
					},
				},
			},
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "invalid chart type",
			"data":    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    chartData,
	})
}

// Handler_CustomTaskExecutionHistory 获取自定义任务执行历史
func Handler_CustomTaskExecutionHistory(c fiber.Ctx) error {
	// 获取图表类型参数
	chartType := c.Query("chartType", "list")

	var chartData interface{}

	switch chartType {
	case "list":
		// 列表模式 - 返回分页数据
		// 获取分页参数
		pageStr := c.Query("page", "1")
		page := 1
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}

		pageSizeStr := c.Query("perPage", "20")
		pageSize := 20
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
		if pageSize > 100 {
			pageSize = 100
		}

		// 获取筛选参数
		taskID := c.Query("task_id")
		status := c.Query("status")
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")

		history, total, err := getCustomTaskExecutionHistory(page, pageSize, taskID, status, startDate, endDate)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": "failed to get execution history",
				"data":    nil,
			})
		}

		chartData = fiber.Map{
			"items":   history,
			"total":   total,
			"page":    page,
			"perPage": pageSize,
		}
	case "timeline":
		// 时间线图
		history, _, _ := getCustomTaskExecutionHistory(1, 50, "", "", "", "")
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "time",
				"axisLabel": map[string]interface{}{
					"formatter": "{MM-dd HH:mm}",
				},
			},
			"yAxis": map[string]interface{}{
				"type": "category",
				"data": func() []string {
					var taskNames []string
					for _, item := range history {
						taskNames = append(taskNames, item.TaskName)
					}
					return taskNames
				}(),
			},
			"series": []map[string]interface{}{
				{
					"name": "执行历史",
					"type": "scatter",
					"data": func() []map[string]interface{} {
						var data []map[string]interface{}
						for _, item := range history {
							color := "#67C23A" // 成功
							switch item.Status {
							case "failed":
								color = "#F56C6C"
							case "running":
								color = "#E6A23C"
							}
							data = append(data, map[string]interface{}{
								"value": []interface{}{item.StartTime.Format("2006-01-02 15:04:05"), item.TaskName},
								"itemStyle": map[string]interface{}{
									"color": color,
								},
							})
						}
						return data
					}(),
				},
			},
		}
	case "status_distribution":
		// 状态分布图
		history, _, _ := getCustomTaskExecutionHistory(1, 1000, "", "", "", "")
		statusCount := make(map[string]int)
		for _, item := range history {
			statusCount[item.Status]++
		}

		var statuses []string
		var counts []int
		for status, count := range statusCount {
			statuses = append(statuses, status)
			counts = append(counts, count)
		}

		chartData = map[string]interface{}{
			"series": []map[string]interface{}{
				{
					"name":   "状态分布",
					"type":   "pie",
					"radius": "50%",
					"data": func() []map[string]interface{} {
						var data []map[string]interface{}
						for i, status := range statuses {
							data = append(data, map[string]interface{}{
								"name":  status,
								"value": counts[i],
							})
						}
						return data
					}(),
					"label": map[string]interface{}{
						"show":      true,
						"formatter": "{b}: {c}({d}%)",
					},
				},
			},
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "invalid chart type",
			"data":    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    chartData,
	})
}

// Handler_CustomTaskTopPerformers 获取自定义任务性能排行
func Handler_CustomTaskTopPerformers(c fiber.Ctx) error {
	// 获取图表类型参数
	chartType := c.Query("chartType", "ranking")

	var chartData interface{}

	switch chartType {
	case "ranking":
		// 排行榜模式 - 返回原始数据
		limitStr := c.Query("limit", "10")
		limit := 10
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}

		if limit > 50 {
			limit = 50
		}

		performers := getCustomTaskTopPerformers(limit)
		chartData = performers
	case "performance_bar":
		// 性能柱状图
		limitStr := c.Query("limit", "10")
		limit := 10
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}

		if limit > 50 {
			limit = 50
		}

		performers := getCustomTaskTopPerformers(limit)

		var taskNames []string
		var successRates []float64
		var avgDurations []float64

		for _, performer := range performers {
			taskNames = append(taskNames, performer.TaskName)
			successRates = append(successRates, performer.SuccessRate)
			avgDurations = append(avgDurations, performer.AvgDuration)
		}

		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": taskNames,
			},
			"yAxis": []map[string]interface{}{
				{
					"type": "value",
					"name": "成功率(%)",
					"max":  100,
				},
				{
					"type": "value",
					"name": "平均执行时间(秒)",
				},
			},
			"series": []map[string]interface{}{
				{
					"name":       "成功率",
					"type":       "bar",
					"yAxisIndex": 0,
					"data":       successRates,
					"itemStyle": map[string]interface{}{
						"color": "#67C23A",
						"normal": map[string]interface{}{
							"label": map[string]interface{}{
								"show":     true,
								"position": "top",
								"textStyle": map[string]interface{}{
									"color":    "#67C23A",
									"fontSize": 16,
								},
							},
						},
					},
				},
				{
					"name":       "平均执行时间",
					"type":       "line",
					"yAxisIndex": 1,
					"data":       avgDurations,
					"itemStyle": map[string]interface{}{
						"color": "#409EFF",
						"normal": map[string]interface{}{
							"label": map[string]interface{}{
								"show":     true,
								"position": "top",
								"textStyle": map[string]interface{}{
									"color":    "#409EFF",
									"fontSize": 16,
								},
							},
						},
					},
				},
			},
		}
	case "success_rate_pie":
		// 成功率饼图
		limitStr := c.Query("limit", "10")
		limit := 10
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}

		if limit > 50 {
			limit = 50
		}

		performers := getCustomTaskTopPerformers(limit)

		chartData = map[string]interface{}{
			"series": []map[string]interface{}{
				{
					"name":   "任务成功率",
					"type":   "pie",
					"radius": "50%",
					"data": func() []map[string]interface{} {
						var data []map[string]interface{}
						for _, performer := range performers {
							data = append(data, map[string]interface{}{
								"name":  performer.TaskName,
								"value": performer.SuccessRate,
							})
						}
						return data
					}(),
					"label": map[string]interface{}{
						"show":      true,
						"formatter": "{b}: {c}%",
					},
				},
			},
		}
	case "execution_count_bar":
		// 执行次数柱状图
		limitStr := c.Query("limit", "10")
		limit := 10
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}

		if limit > 50 {
			limit = 50
		}

		performers := getCustomTaskTopPerformers(limit)

		var taskNames []string
		var executionCounts []int64

		for _, performer := range performers {
			taskNames = append(taskNames, performer.TaskName)
			executionCounts = append(executionCounts, performer.ExecutionCount)
		}

		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": taskNames,
			},
			"yAxis": map[string]interface{}{
				"type": "value",
				"name": "执行次数",
			},
			"series": []map[string]interface{}{
				{
					"name": "执行次数",
					"type": "bar",
					"data": executionCounts,
					"itemStyle": map[string]interface{}{
						"color": "#E6A23C",
					},
				},
			},
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "invalid chart type",
			"data":    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    chartData,
	})
}

// getCustomTaskDetailedStats 获取自定义任务详细统计
func getCustomTaskDetailedStats() CustomTaskDetailedStats {
	var stats CustomTaskDetailedStats

	// 统计任务总数
	config.DB.Model(&task.CustomTask{}).Count(&stats.TotalTasks)

	// 统计活跃任务数
	config.DB.Model(&task.CustomTask{}).Where("status = ?", "active").Count(&stats.ActiveTasks)

	// 统计执行总数
	config.DB.Model(&task.CustomTaskExecution{}).Count(&stats.TotalExecutions)

	// 统计成功执行数
	config.DB.Model(&task.CustomTaskExecution{}).Where("status = ?", "completed").Count(&stats.SuccessfulExecutions)

	// 统计失败执行数
	config.DB.Model(&task.CustomTaskExecution{}).Where("status = ?", "failed").Count(&stats.FailedExecutions)

	// 统计运行中执行数
	config.DB.Model(&task.CustomTaskExecution{}).Where("status = ?", "running").Count(&stats.RunningExecutions)

	// 计算平均执行时间
	var avgTime float64
	config.DB.Table("custom_task_executions").
		Select("AVG(CASE WHEN end_time IS NOT NULL THEN TIMESTAMPDIFF(SECOND, start_time, end_time) ELSE 0 END)").
		Where("end_time IS NOT NULL").
		Scan(&avgTime)
	stats.AvgExecutionTime = avgTime

	// 计算成功率
	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(stats.SuccessfulExecutions) / float64(stats.TotalExecutions) * 100
	}

	return stats
}

// getCustomTaskExecutionHistory 获取自定义任务执行历史
func getCustomTaskExecutionHistory(page, pageSize int, taskID, status, startDate, endDate string) ([]CustomTaskExecutionHistory, int64, error) {
	var history []CustomTaskExecutionHistory
	var total int64

	query := config.DB.Table("custom_task_executions e").
		Select(`
			e.task_id,
			t.name as task_name,
			e.status,
			e.start_time,
			e.end_time,
			CASE WHEN e.end_time IS NOT NULL THEN TIMESTAMPDIFF(SECOND, e.start_time, e.end_time) ELSE 0 END as duration,
			e.error_msg
		`).
		Joins("JOIN custom_tasks t ON e.task_id = t.id")

	// 应用筛选条件
	if taskID != "" {
		query = query.Where("e.task_id = ?", taskID)
	}
	if status != "" {
		query = query.Where("e.status = ?", status)
	}
	if startDate != "" {
		query = query.Where("DATE(e.start_time) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(e.start_time) <= ?", endDate)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("e.start_time DESC").Offset(offset).Limit(pageSize).Find(&history).Error; err != nil {
		return nil, 0, err
	}

	return history, total, nil
}

// getCustomTaskTopPerformers 获取自定义任务性能排行
func getCustomTaskTopPerformers(limit int) []TaskPerformanceData {
	var performers []TaskPerformanceData

	config.DB.Table("custom_task_executions e").
		Select(`
			t.name as task_name,
			AVG(CASE WHEN e.end_time IS NOT NULL THEN TIMESTAMPDIFF(SECOND, e.start_time, e.end_time) ELSE 0 END) as avg_duration,
			(COUNT(CASE WHEN e.status = 'completed' THEN 1 END) * 100.0 / COUNT(*)) as success_rate,
			COUNT(*) as execution_count
		`).
		Joins("JOIN custom_tasks t ON e.task_id = t.id").
		Group("t.id, t.name").
		Having("execution_count >= 1").
		Order("success_rate DESC, avg_duration ASC").
		Limit(limit).
		Find(&performers)

	return performers
}
