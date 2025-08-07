// Package router CMDB中主机记录相关API接口
package route

import (
	"saurfang/internal/config"
	"saurfang/internal/handler/cmdbhandler"
	"saurfang/internal/models/autosync"
	"saurfang/internal/models/gamegroup"
	"saurfang/internal/models/gamehost"
	"saurfang/internal/repository/base"

	"github.com/gofiber/fiber/v3"
)

type CMDBRouteModule struct {
	Namespace string // 路由组
	Comment   string //说明
}

func (c *CMDBRouteModule) Info() (namespace string, comment string) {
	return c.Namespace, c.Comment
}
func (c *CMDBRouteModule) RegisterRoutesModule(r *fiber.App) {
	// cmdb路由组
	cmdbRouter := r.Group(c.Namespace)
	/*
		主机记录
	*/
	//cmdbHostService := cmdbservice.NewHostsService(config.DB)
	//cmdbHostHandler := cmdbhandler.NewHostHandler(cmdbHostService)
	cmdbHostHandler := cmdbhandler.HostHandler{BaseGormRepository: base.BaseGormRepository[gamehost.Hosts]{DB: config.DB}}
	cmdbRouter.Get("/host/list", cmdbHostHandler.Handler_ListHosts)
	// 分页显示主机记录
	cmdbRouter.Get("/host/listPerPage", cmdbHostHandler.Handler_ListHostsPerPage)
	// 创建主机记录
	cmdbRouter.Post("/host/create", cmdbHostHandler.Handler_CreateHost)
	// 快速保存
	cmdbRouter.Post("/host/quickSave", cmdbHostHandler.Handler_QuickSave)
	// 删除主机记录
	cmdbRouter.Delete("/host/delete/:id", cmdbHostHandler.Handler_DeleteHost)
	// 批量删除主机记录
	cmdbRouter.Delete("/host/batchDelete/:ids", cmdbHostHandler.Handler_BatchDeleteHosts)
	// 更新主机记录
	cmdbRouter.Put("/host/update/:id", cmdbHostHandler.Handler_UpdateHost)
	// 重新分配归属组
	cmdbRouter.Put("/host/regroup/:id/:greoup_id", cmdbHostHandler.Handler_ReGroup)
	/*
		归属组
	*/
	// 归属组路由api
	//cmdbGroupService := cmdbservice.NewGroupsService(config.DB)
	//cmdbGroupHandler := cmdbhandler.NewGroupHandler(cmdbGroupService)
	cmdbGroupHandler := cmdbhandler.GroupHandler{BaseGormRepository: base.BaseGormRepository[gamegroup.Groups]{DB: config.DB}}
	// 创建新归属组
	cmdbRouter.Post("/group/create", cmdbGroupHandler.Handler_CreateNewGroup)
	//删除组
	cmdbRouter.Delete("/group/delete/:id", cmdbGroupHandler.Handler_DeleteGroup)
	// 批量删除组
	cmdbRouter.Delete("/group/batchDelete", cmdbGroupHandler.Handler_BatchDeleteGroups)
	// 更新组信息
	cmdbRouter.Put("/group/update/:id", cmdbGroupHandler.Handler_UpdateGroup)
	// 展示归属组
	cmdbRouter.Get("/group/list", cmdbGroupHandler.Handler_ListGroups)
	// Amis前端组ID转换为组名称
	cmdbRouter.Get("/group/idtoname", cmdbGroupHandler.Handler_GroupIdToName)
	// 新增主机选择归属组
	cmdbRouter.Get("/group/select", cmdbGroupHandler.Handler_SelectGroupForHost)
	/*
		云主机自动同步
	*/

	syncHandler := cmdbhandler.AutoSyncHandler{BaseGormRepository: base.BaseGormRepository[autosync.AutoSync]{DB: config.DB}}
	cmdbRouter.Post("/sync/create", syncHandler.Handler_CreateAutoSyncConfig)
	cmdbRouter.Delete("/sync/delete/:id", syncHandler.Handler_DeleteAutoSyncConfig)
	cmdbRouter.Put("/sync/update/:id", syncHandler.Handler_UpdateAutoSyncConfig)
	cmdbRouter.Get("/sync/list", syncHandler.Handler_ShowAutoSyncConfig)
	cmdbRouter.Get("/sync/select", syncHandler.Handler_AutoSyncConfigSelect)
	cmdbRouter.Post("/sync/run", syncHandler.Handler_AutoSync)
}
func init() {
	RegisterRoutesModule(&CMDBRouteModule{Namespace: "/api/v1/cmdb", Comment: "资源管理"})
}
