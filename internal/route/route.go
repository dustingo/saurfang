// Package router CMDB中主机记录相关API接口
package route

import (
	"saurfang/internal/config"
	"saurfang/internal/handler/cmdbhandler"
	"saurfang/internal/service/cmdbservice.go"

	"github.com/gofiber/fiber/v2"
)

func CMDBRouter(app *fiber.App) *fiber.Router {
	// 主机路由api
	cmdbHostService := cmdbservice.NewHostsService(config.DB)
	cmdbHostHandler := cmdbhandler.NewHostHandler(cmdbHostService)
	cmdbRouter := app.Group("/api/v1/cmdb")
	// 列出主机记录
	cmdbRouter.Get("/host/list", cmdbHostHandler.Handler_ListHosts)
	// 分页显示主机记录
	cmdbRouter.Get("/host/listPerPage", cmdbHostHandler.Handler_ListHostsPerPage)
	// 创建主机记录
	cmdbRouter.Post("/host/create", cmdbHostHandler.Handler_CreateHost)
	// 快速保存
	cmdbRouter.Post("host/quickSave", cmdbHostHandler.Handler_QuickSave)
	// 删除主机记录
	cmdbRouter.Delete("/host/delete/:id", cmdbHostHandler.Handler_DeleteHost)
	// 批量删除主机记录
	cmdbRouter.Delete("/host/batchDelete/:ids", cmdbHostHandler.Handler_BatchDeleteHosts)
	// 更新主机记录
	cmdbRouter.Put("/host/update/:id", cmdbHostHandler.Handler_UpdateHost)
	// 重新分配归属组
	cmdbRouter.Put("/host/regroup/:id/:greoup_id", cmdbHostHandler.Handler_ReGroup)
	// 归属组路由api
	cmdbGroupService := cmdbservice.NewGroupsService(config.DB)
	return &cmdbRouter
}
