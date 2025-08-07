package route

import (
	"saurfang/internal/handler/boardhandler"

	"github.com/gofiber/fiber/v3"
)

type DashboardRouteModule struct {
	Namespace string // 路由组
	Comment   string //说明
}

func (d *DashboardRouteModule) Info() (namespace string, comment string) {
	return d.Namespace, d.Comment
}
func (d *DashboardRouteModule) RegisterRoutesModule(r *fiber.App) {
	dashboardRoute := r.Group(d.Namespace)
	/*
		资源统计
	*/
	dashboardRoute.Get("/resource", boardhandler.Handler_TotalResource)

	/*
		用户登录
	*/
	dashboardRoute.Get("/login/records", boardhandler.Handler_LoginRecords)

	/*
		Dashboard综合统计
	*/
	dashboardRoute.Get("/stats", boardhandler.Handler_DashboardStats)

	/*
		自定义任务图表统计
	*/
	dashboardRoute.Get("/custom-task-charts", boardhandler.Handler_CustomTaskCharts)

	/*
		自定义任务详细统计
	*/
	dashboardRoute.Get("/custom-task-stats", boardhandler.Handler_CustomTaskDetailedStats)
	dashboardRoute.Get("/custom-task-history", boardhandler.Handler_CustomTaskExecutionHistory)
	dashboardRoute.Get("/custom-task-performers", boardhandler.Handler_CustomTaskTopPerformers)
}
func init() {
	RegisterRoutesModule(&DashboardRouteModule{Namespace: "/api/v1/dashboard", Comment: "看板路由"})
}
