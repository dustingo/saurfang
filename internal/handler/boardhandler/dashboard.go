package boardhandler

import (
	"saurfang/internal/config"
	"saurfang/internal/models/task"
	"time"

	"github.com/gofiber/fiber/v3"
)

// DashboardStats dashboard统计数据
type DashboardStats struct {
	ResourceStats    ResourceStats     `json:"resource_stats"`
	TaskStats        TaskStats         `json:"task_stats"`
	ClusterStats     NomadClusterStats `json:"cluster_stats"`
	RecentActivities []RecentActivity  `json:"recent_activities"`
}

// ResourceStats 资源统计
type ResourceStats struct {
	Channels int64 `json:"channels"`
	Hosts    int64 `json:"hosts"`
	Groups   int64 `json:"groups"`
	Games    int64 `json:"games"`
	Users    int64 `json:"users"`
}

// TaskStats 任务统计
type TaskStats struct {
	TotalCronJobs      int64 `json:"total_cron_jobs"`
	ActiveCronJobs     int64 `json:"active_cron_jobs"`
	TotalCustomTasks   int64 `json:"total_custom_tasks"`
	RunningCustomTasks int64 `json:"running_custom_tasks"`
	CompletedTasks     int64 `json:"completed_tasks"`
	FailedTasks        int64 `json:"failed_tasks"`
}

// CustomTaskChartData 自定义任务图表数据
type CustomTaskChartData struct {
	ExecutionTrend    []ExecutionTrendData    `json:"execution_trend"`     // 执行趋势
	ScriptTypeDist    []ScriptTypeDistData    `json:"script_type_dist"`    // 脚本类型分布
	ExecutionStatus   []ExecutionStatusData   `json:"execution_status"`    // 执行状态统计
	TaskPerformance   []TaskPerformanceData   `json:"task_performance"`    // 任务性能统计
	HostExecutionDist []HostExecutionDistData `json:"host_execution_dist"` // 主机执行分布
}

// ExecutionTrendData 执行趋势数据
type ExecutionTrendData struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ScriptTypeDistData 脚本类型分布数据
type ScriptTypeDistData struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// ExecutionStatusData 执行状态数据
type ExecutionStatusData struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// TaskPerformanceData 任务性能数据
type TaskPerformanceData struct {
	TaskName       string  `json:"task_name"`
	AvgDuration    float64 `json:"avg_duration"`
	SuccessRate    float64 `json:"success_rate"`
	ExecutionCount int64   `json:"execution_count"`
}

// HostExecutionDistData 主机执行分布数据
type HostExecutionDistData struct {
	HostName string `json:"host_name"`
	Count    int64  `json:"count"`
}

// ClusterStats 集群统计
// type NomadClusterStats struct {
// 	TotalNodes   int64 `json:"total_nodes"`
// 	OnlineNodes  int64 `json:"online_nodes"`
// 	OfflineNodes int64 `json:"offline_nodes"`
// 	TotalJobs    int64 `json:"total_jobs"`
// 	RunningJobs  int64 `json:"running_jobs"`
// 	StoppedJobs  int64 `json:"stopped_jobs"`
// }

// RecentActivity 最近活动
type RecentActivity struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func Handler_DashboardStats(c fiber.Ctx) error {
	// 获取图表类型参数
	chartType := c.Query("chartType", "pie")

	var chartData map[string]interface{}

	switch chartType {
	case "pie":
		// 饼图数据 - 任务统计
		taskStats := getTaskStats()
		chartData = map[string]interface{}{
			"series": []map[string]interface{}{
				{
					"type":   "pie",
					"radius": "50%",
					"data": []map[string]interface{}{
						{"name": "定时任务", "value": taskStats.TotalCronJobs},
						{"name": "活跃定时任务", "value": taskStats.ActiveCronJobs},
						{"name": "自定义任务", "value": taskStats.TotalCustomTasks},
						{"name": "运行中任务", "value": taskStats.RunningCustomTasks},
						{"name": "已完成任务", "value": taskStats.CompletedTasks},
						{"name": "失败任务", "value": taskStats.FailedTasks},
					},
					"label": map[string]interface{}{
						"show":      true,
						"formatter": "{b}: {c}({d}%)",
					},
				},
			},
		}
	case "line":
		// 折线图数据 - 集群统计
		clusterStats := getRealClusterStats()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": []string{"总节点", "在线节点", "离线节点", "总任务", "运行中任务", "停止任务"},
			},
			"yAxis": map[string]interface{}{
				"type": "value",
			},
			"series": []map[string]interface{}{
				{
					"name": "数量",
					"type": "line",
					"data": []int64{clusterStats.TotalNodes, clusterStats.OnlineNodes, clusterStats.OfflineNodes, clusterStats.TotalJobs, clusterStats.RunningJobs, clusterStats.StoppedJobs},
					"itemStyle": map[string]interface{}{
						"normal": map[string]interface{}{
							"label": map[string]interface{}{
								"show":     true,
								"position": "top",
								"textStyle": map[string]interface{}{
									"color":    "black",
									"fontSize": 16,
								},
							},
						},
					},
				},
			},
		}
	case "custom_task_trend":
		// 自定义任务执行趋势图
		trendData := getCustomTaskExecutionTrend()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": func() []string {
					var dates []string
					for _, item := range trendData {
						dates = append(dates, item.Date)
					}
					return dates
				}(),
			},
			"yAxis": map[string]interface{}{
				"type": "value",
			},
			"series": []map[string]interface{}{
				{
					"name": "执行次数",
					"type": "line",
					"data": func() []int64 {
						var counts []int64
						for _, item := range trendData {
							counts = append(counts, item.Count)
						}
						return counts
					}(),
					"smooth": true,
				},
			},
		}
	case "script_type_dist":
		// 脚本类型分布图
		scriptTypeData := getScriptTypeDistribution()
		chartData = map[string]interface{}{
			"series": []map[string]interface{}{
				{
					"type":   "pie",
					"radius": "50%",
					"data": func() []map[string]interface{} {
						var data []map[string]interface{}
						for _, item := range scriptTypeData {
							data = append(data, map[string]interface{}{
								"name":  item.Type,
								"value": item.Count,
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
	case "execution_status":
		// 执行状态统计图
		statusData := getExecutionStatusStats()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": func() []string {
					var statuses []string
					for _, item := range statusData {
						statuses = append(statuses, item.Status)
					}
					return statuses
				}(),
			},
			"yAxis": map[string]interface{}{
				"type": "value",
			},
			"series": []map[string]interface{}{
				{
					"name": "数量",
					"type": "bar",
					"data": func() []int64 {
						var counts []int64
						for _, item := range statusData {
							counts = append(counts, item.Count)
						}
						return counts
					}(),
				},
			},
		}
	case "task_performance":
		// 任务性能统计图
		performanceData := getTaskPerformanceStats()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": func() []string {
					var names []string
					for _, item := range performanceData {
						names = append(names, item.TaskName)
					}
					return names
				}(),
			},
			"yAxis": []map[string]interface{}{
				{
					"type": "value",
					"name": "平均执行时间(秒)",
				},
				{
					"type": "value",
					"name": "成功率(%)",
					"max":  100,
				},
			},
			"series": []map[string]interface{}{
				{
					"name":       "平均执行时间",
					"type":       "bar",
					"yAxisIndex": 0,
					"data": func() []float64 {
						var durations []float64
						for _, item := range performanceData {
							durations = append(durations, item.AvgDuration)
						}
						return durations
					}(),
				},
				{
					"name":       "成功率",
					"type":       "line",
					"yAxisIndex": 1,
					"data": func() []float64 {
						var rates []float64
						for _, item := range performanceData {
							rates = append(rates, item.SuccessRate)
						}
						return rates
					}(),
				},
			},
		}
	case "host_execution_dist":
		// 主机执行分布图
		hostData := getHostExecutionDistribution()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": func() []string {
					var hosts []string
					for _, item := range hostData {
						hosts = append(hosts, item.HostName)
					}
					return hosts
				}(),
			},
			"yAxis": map[string]interface{}{
				"type": "value",
			},
			"series": []map[string]interface{}{
				{
					"name": "执行次数",
					"type": "bar",
					"data": func() []int64 {
						var counts []int64
						for _, item := range hostData {
							counts = append(counts, item.Count)
						}
						return counts
					}(),
				},
			},
		}
	default:
		// 默认返回完整的统计数据
		var stats DashboardStats
		stats.ResourceStats = getResourceStats()
		stats.TaskStats = getTaskStats()
		stats.ClusterStats = getRealClusterStats()
		stats.RecentActivities = getRecentActivities()

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  0,
			"message": "success",
			"data":    stats,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    chartData,
	})
}

// Handler_CustomTaskCharts 自定义任务图表数据处理器
func Handler_CustomTaskCharts(c fiber.Ctx) error {
	// 获取图表类型参数
	chartType := c.Query("chartType", "all")

	var chartData interface{}

	switch chartType {
	case "execution_trend":
		// 自定义任务执行趋势图
		trendData := getCustomTaskExecutionTrend()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": func() []string {
					var dates []string
					for _, item := range trendData {
						dates = append(dates, item.Date)
					}
					return dates
				}(),
			},
			"yAxis": map[string]interface{}{
				"type": "value",
			},
			"series": []map[string]interface{}{
				{
					"name": "执行次数",
					"type": "line",
					"data": func() []int64 {
						var counts []int64
						for _, item := range trendData {
							counts = append(counts, item.Count)
						}
						return counts
					}(),
					"smooth": true,
				},
			},
		}
	case "script_type_dist":
		// 脚本类型分布图
		scriptTypeData := getScriptTypeDistribution()
		chartData = map[string]interface{}{
			"series": []map[string]interface{}{
				{
					"name":   "脚本类型分布",
					"type":   "pie",
					"radius": "50%",
					"data": func() []map[string]interface{} {
						var data []map[string]interface{}
						for _, item := range scriptTypeData {
							data = append(data, map[string]interface{}{
								"name":  item.Type,
								"value": item.Count,
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
	case "execution_status":
		// 执行状态统计图
		statusData := getExecutionStatusStats()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": func() []string {
					var statuses []string
					for _, item := range statusData {
						statuses = append(statuses, item.Status)
					}
					return statuses
				}(),
			},
			"yAxis": map[string]interface{}{
				"type": "value",
			},
			"series": []map[string]interface{}{
				{
					"name": "数量",
					"type": "bar",
					"data": func() []int64 {
						var counts []int64
						for _, item := range statusData {
							counts = append(counts, item.Count)
						}
						return counts
					}(),
				},
			},
		}
	case "task_performance":
		// 任务性能统计图
		performanceData := getTaskPerformanceStats()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": func() []string {
					var names []string
					for _, item := range performanceData {
						names = append(names, item.TaskName)
					}
					return names
				}(),
			},
			"yAxis": []map[string]interface{}{
				{
					"type": "value",
					"name": "平均执行时间(秒)",
				},
				{
					"type": "value",
					"name": "成功率(%)",
					"max":  100,
				},
			},
			"series": []map[string]interface{}{
				{
					"name":       "平均执行时间",
					"type":       "bar",
					"yAxisIndex": 0,
					"data": func() []float64 {
						var durations []float64
						for _, item := range performanceData {
							durations = append(durations, item.AvgDuration)
						}
						return durations
					}(),
				},
				{
					"name":       "成功率",
					"type":       "line",
					"yAxisIndex": 1,
					"data": func() []float64 {
						var rates []float64
						for _, item := range performanceData {
							rates = append(rates, item.SuccessRate)
						}
						return rates
					}(),
				},
			},
		}
	case "host_execution_dist":
		// 主机执行分布图
		hostData := getHostExecutionDistribution()
		chartData = map[string]interface{}{
			"xAxis": map[string]interface{}{
				"type": "category",
				"data": func() []string {
					var hosts []string
					for _, item := range hostData {
						hosts = append(hosts, item.HostName)
					}
					return hosts
				}(),
			},
			"yAxis": map[string]interface{}{
				"type": "value",
			},
			"series": []map[string]interface{}{
				{
					"name": "执行次数",
					"type": "bar",
					"data": func() []int64 {
						var counts []int64
						for _, item := range hostData {
							counts = append(counts, item.Count)
						}
						return counts
					}(),
				},
			},
		}
	case "all":
		// 返回所有图表数据
		var allChartData CustomTaskChartData
		allChartData.ExecutionTrend = getCustomTaskExecutionTrend()
		allChartData.ScriptTypeDist = getScriptTypeDistribution()
		allChartData.ExecutionStatus = getExecutionStatusStats()
		allChartData.TaskPerformance = getTaskPerformanceStats()
		allChartData.HostExecutionDist = getHostExecutionDistribution()
		chartData = allChartData
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

func getResourceStats() ResourceStats {
	var stats ResourceStats

	// 统计渠道数量
	config.DB.Model(&struct{}{}).Table("channels").Count(&stats.Channels)

	// 统计主机数量
	config.DB.Model(&struct{}{}).Table("hosts").Count(&stats.Hosts)

	// 统计游戏组数量
	config.DB.Model(&struct{}{}).Table("groups").Count(&stats.Groups)

	// 统计游戏服务器数量
	config.DB.Model(&struct{}{}).Table("games").Count(&stats.Games)

	// 统计用户数量
	config.DB.Model(&struct{}{}).Table("users").Count(&stats.Users)

	return stats
}

func getTaskStats() TaskStats {
	var stats TaskStats

	// 统计定时任务
	config.DB.Model(&task.CronJobs{}).Count(&stats.TotalCronJobs)
	config.DB.Model(&task.CronJobs{}).Where("task_status = ?", 0).Count(&stats.ActiveCronJobs)

	// 统计自定义任务
	config.DB.Model(&task.CustomTask{}).Count(&stats.TotalCustomTasks)
	config.DB.Model(&task.CustomTaskExecution{}).Where("status = ?", "running").Count(&stats.RunningCustomTasks)

	// 统计任务执行结果
	config.DB.Model(&task.CustomTaskExecution{}).Where("status = ?", "completed").Count(&stats.CompletedTasks)
	config.DB.Model(&task.CustomTaskExecution{}).Where("status = ?", "failed").Count(&stats.FailedTasks)

	return stats
}

func getRecentActivities() []RecentActivity {
	var activities []RecentActivity

	// 获取最近的登录记录
	var loginRecords []struct {
		ID        uint      `json:"id"`
		Username  string    `json:"username"`
		LastLogin time.Time `json:"last_login"`
		ClientIP  string    `json:"client_ip"`
	}

	config.DB.Table("login_records").Order("last_login desc").Limit(5).Find(&loginRecords)

	for _, record := range loginRecords {
		activities = append(activities, RecentActivity{
			ID:        record.ID,
			Type:      "login",
			Message:   "用户 " + record.Username + " 登录系统",
			Timestamp: record.LastLogin,
		})
	}

	// 获取最近的任务执行记录
	var taskExecutions []struct {
		ID        uint      `json:"id"`
		TaskName  string    `json:"task_name"`
		Status    string    `json:"status"`
		StartTime time.Time `json:"start_time"`
	}

	config.DB.Table("custom_task_executions").Order("start_time desc").Limit(5).Find(&taskExecutions)

	for _, execution := range taskExecutions {
		activities = append(activities, RecentActivity{
			ID:        execution.ID,
			Type:      "task",
			Message:   "任务 " + execution.TaskName + " " + execution.Status,
			Timestamp: execution.StartTime,
		})
	}

	return activities
}

// getCustomTaskExecutionTrend 获取自定义任务执行趋势
func getCustomTaskExecutionTrend() []ExecutionTrendData {
	var trendData []ExecutionTrendData

	// 获取最近7天的执行趋势
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var count int64

		config.DB.Model(&task.CustomTaskExecution{}).
			Where("DATE(start_time) = ?", date).
			Count(&count)

		trendData = append(trendData, ExecutionTrendData{
			Date:  date,
			Count: count,
		})
	}

	return trendData
}

// getScriptTypeDistribution 获取脚本类型分布
func getScriptTypeDistribution() []ScriptTypeDistData {
	var scriptTypeData []ScriptTypeDistData

	rows, err := config.DB.Model(&task.CustomTask{}).
		Select("script_type, COUNT(*) as count").
		Group("script_type").
		Rows()

	if err != nil {
		return scriptTypeData
	}
	defer rows.Close()

	for rows.Next() {
		var scriptType string
		var count int64
		if err := rows.Scan(&scriptType, &count); err == nil {
			scriptTypeData = append(scriptTypeData, ScriptTypeDistData{
				Type:  scriptType,
				Count: count,
			})
		}
	}

	return scriptTypeData
}

// getExecutionStatusStats 获取执行状态统计
func getExecutionStatusStats() []ExecutionStatusData {
	var statusData []ExecutionStatusData

	rows, err := config.DB.Model(&task.CustomTaskExecution{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Rows()

	if err != nil {
		return statusData
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err == nil {
			statusData = append(statusData, ExecutionStatusData{
				Status: status,
				Count:  count,
			})
		}
	}

	return statusData
}

// getTaskPerformanceStats 获取任务性能统计
func getTaskPerformanceStats() []TaskPerformanceData {
	var performanceData []TaskPerformanceData

	rows, err := config.DB.Table("custom_task_executions").
		Select(`
			t.name as task_name,
			AVG(CASE WHEN e.end_time IS NOT NULL THEN TIMESTAMPDIFF(SECOND, e.start_time, e.end_time) ELSE 0 END) as avg_duration,
			(COUNT(CASE WHEN e.status = 'completed' THEN 1 END) * 100.0 / COUNT(*)) as success_rate,
			COUNT(*) as execution_count
		`).
		Joins("JOIN custom_tasks t ON e.task_id = t.id").
		Group("t.id, t.name").
		Order("execution_count DESC").
		Limit(10).
		Rows()

	if err != nil {
		return performanceData
	}
	defer rows.Close()

	for rows.Next() {
		var taskName string
		var avgDuration, successRate float64
		var executionCount int64

		if err := rows.Scan(&taskName, &avgDuration, &successRate, &executionCount); err == nil {
			performanceData = append(performanceData, TaskPerformanceData{
				TaskName:       taskName,
				AvgDuration:    avgDuration,
				SuccessRate:    successRate,
				ExecutionCount: executionCount,
			})
		}
	}

	return performanceData
}

// getHostExecutionDistribution 获取主机执行分布
func getHostExecutionDistribution() []HostExecutionDistData {
	var hostData []HostExecutionDistData

	// 从任务的目标主机中统计执行分布
	rows, err := config.DB.Table("custom_tasks").
		Select("target_hosts, COUNT(*) as count").
		Where("target_hosts != ''").
		Group("target_hosts").
		Order("count DESC").
		Limit(10).
		Rows()

	if err != nil {
		return hostData
	}
	defer rows.Close()

	for rows.Next() {
		var targetHosts string
		var count int64

		if err := rows.Scan(&targetHosts, &count); err == nil {
			// 处理多个主机的情况
			hosts := func() []string {
				if targetHosts == "" {
					return []string{}
				}
				// 简单的逗号分割，实际可能需要更复杂的解析
				return []string{targetHosts}
			}()

			for _, host := range hosts {
				hostData = append(hostData, HostExecutionDistData{
					HostName: host,
					Count:    count,
				})
			}
		}
	}

	return hostData
}
