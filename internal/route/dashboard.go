package route

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/handler/boardhandler"
)

type DashboardRouteModule struct {
	Namespace string // 路由组
	Comment   string //说明
}

func (d *DashboardRouteModule) Info() (namespace string, comment string) {
	namespace = d.Namespace
	comment = d.Comment
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
}
func init() {
	RegisterRoutesModule(&DashboardRouteModule{Namespace: "/api/v1/dashboard", Comment: "看板路由"})
}
