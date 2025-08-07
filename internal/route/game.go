package route

import (
	"os"
	"saurfang/internal/config"
	"saurfang/internal/handler/gamehandler"
	"saurfang/internal/models/gamechannel"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/repository/base"

	"github.com/gofiber/fiber/v3"
)

type GameRouteModule struct {
	Namespace string
	Comment   string
}

func (g *GameRouteModule) Info() (namespace string, comment string) {
	return g.Namespace, g.Comment
}

func (g *GameRouteModule) RegisterRoutesModule(r *fiber.App) {
	gameRouter := r.Group(g.Namespace)
	/*
		渠道
	*/
	channelHandler := gamehandler.ChannelHandler{BaseGormRepository: base.BaseGormRepository[gamechannel.Channels]{DB: config.DB}}
	gameRouter.Post("/channel/create", channelHandler.Handler_CreateChannel)
	gameRouter.Delete("/channel/delete/:id", channelHandler.Handler_DeleteChannel)
	gameRouter.Put("/channel/update/:id", channelHandler.Handler_UpdateChannel)
	gameRouter.Get("/channel/list", channelHandler.Handler_Listhannel)
	gameRouter.Get("/channel/navlist", channelHandler.Handler_AmisNavlist)
	gameRouter.Get("/channel/select", channelHandler.Handler_SelectChannel)
	/*
		逻辑服
	*/
	// nomadcli, _ := config.NewNomadClient()
	logicHandler := gamehandler.LogicServerHandler{
		BaseGormRepository: base.BaseGormRepository[gameserver.Games]{DB: config.DB},
		// NomadJobRepository: base.NomadJobRepository{
		// 	Consul: config.ConsulCli,
		// 	// Ns:     os.Getenv("GAME_CONFIG_NAMESPACE"),
		// 	Nomad: nomadcli,
		// },
	}
	// 创建游戏服
	gameRouter.Post("/logic/create", logicHandler.Handler_CreateLogicServer)
	gameRouter.Delete("/logic/delete", logicHandler.Handler_DeleteLogicServer)
	// gameRouter.Delete("/logic/hosts/delete", logicHandler.Handler_DeleteHostFromLogicServer)
	gameRouter.Put("/logic/update/:id", logicHandler.Handler_UpdateLogicServer)
	gameRouter.Get("/logic/list", logicHandler.Handler_ShowLogicServer)
	//
	gameRouter.Get("/logic/select", logicHandler.Handler_ShowChannelServerList)
	// gameRouter.Get("/logic/detail", logicHandler.Handler_ShowServerDetail)
	//gameRouter.Get("/logic/detail/select", logicHandler.Handler_ShowGameserverByTree)
	gameRouter.Get("/logic/detail/picker", logicHandler.Handler_ShowServerDetailForPicker)
	// gameRouter.Put("/logic/hosts/assign", logicHandler.Handler_AddHostsToLogicServer)
	gameRouter.Get("/logic/config/select", logicHandler.Handler_TreeSelectForSyncServerConfig)
	/*
		逻辑服配置
	*/
	serverconfigHandler := gamehandler.NewServerConfigHandler(config.ConsulCli, os.Getenv("GAME_NOMAD_JOB_NAMESPACE"))
	gameRouter.Post("/config/create", serverconfigHandler.Handler_CreateNomadJob)
	gameRouter.Delete("/config/delete", serverconfigHandler.Handler_DeleteServerConfig)
	gameRouter.Put("/config/update", serverconfigHandler.Handler_UpdateServerConfig)
	gameRouter.Get("/config/list", serverconfigHandler.Handler_ListServerConfig)
	//gameRouter.Get("/config/listByKey/:key", serverconfigHandler.Handler_ListNomadJobByKey)
	gameRouter.Get("/config/:server_id/show", serverconfigHandler.Handler_ListNomadJobByKey)

	/*
		nomad发布配置
	*/
	deployConfigHandler := gamehandler.NewServerConfigHandler(config.ConsulCli, os.Getenv("GAME_NOMAD_DEPLOY_NAMESPACE"))
	gameRouter.Post("/deploy/config/create", deployConfigHandler.Handler_CreateNomadJob)
	gameRouter.Delete("/deploy/config/delete", deployConfigHandler.Handler_DeleteServerConfig)
	gameRouter.Put("/deploy/config/update", deployConfigHandler.Handler_UpdateServerConfig)
	gameRouter.Get("/deploy/config/list", deployConfigHandler.Handler_ListServerConfig)
	gameRouter.Get("/deploy/config/:server_id/show", deployConfigHandler.Handler_ListNomadJobByKey)
}
func init() {
	RegisterRoutesModule(&GameRouteModule{Namespace: "/api/v1/game", Comment: "游戏服管理"})
}
