package route

import (
	"os"
	"saurfang/internal/config"
	"saurfang/internal/handler/gamehandler"
	"saurfang/internal/service/gameservice"

	"github.com/gofiber/fiber/v3"
)

func GameRouter(app *fiber.App) *fiber.Router {
	channelService := gameservice.NewChannelService(config.DB)
	channelHandler := gamehandler.NewHostHandler(channelService)
	/*
		渠道
	*/
	gameRouter := app.Group("/api/v1/game")
	gameRouter.Post("/channel/create", channelHandler.Handler_CreateChannel)
	gameRouter.Delete("/channel/delete/:id", channelHandler.Handler_DeleteChannel)
	gameRouter.Put("/channel/update/:id", channelHandler.Handler_UpdateChannel)
	gameRouter.Get("/channel/list", channelHandler.Handler_Listhannel)
	gameRouter.Get("/channel/navlist", channelHandler.Handler_AmisNavlist)
	gameRouter.Get("/channel/select", channelHandler.Handler_SelectChannel)
	/*
		逻辑服
	*/
	logicservice := gameservice.NewLogicServerService(config.DB, config.Etcd)
	logicHandler := gamehandler.NewLogicServerHandler(logicservice)
	gameRouter.Post("/logic/create", logicHandler.Handler_CreateLogicServer)
	gameRouter.Delete("/logic/delete", logicHandler.Handler_DeleteLogicServer)
	gameRouter.Delete("/logic/hosts/delete", logicHandler.Handler_DeleteHostFromLogicServer)
	gameRouter.Put("/logic/update/:id", logicHandler.Handler_UpdateLogicServer)
	gameRouter.Get("/logic/list", logicHandler.Handler_ShowLogicServer)
	gameRouter.Get("/logic/detail", logicHandler.Handler_ShowServerDetail)
	gameRouter.Get("/logic/detail/select", logicHandler.Handler_ShowGameserverByTree)
	gameRouter.Get("/logic/detail/picker", logicHandler.Handler_ShowServerDetailForPicker)
	gameRouter.Put("/logic/hosts/assign", logicHandler.Handler_AddHostsToLogicServer)
	gameRouter.Get("/logic/config/select", logicHandler.Handler_TreeSelectForSyncServerConfig)
	gameRouter.Get("/logic/process/list", logicHandler.Handler_ShowGameProcesses)
	gameRouter.Put("/logic/start", logicHandler.Handler_ExecGameops)
	gameRouter.Put("/logic/stop", logicHandler.Handler_ExecGameops)
	gameRouter.Put("/logic/batch/start", logicHandler.Handler_BatchExecgameops)
	gameRouter.Put("/logic/batch/stop", logicHandler.Handler_BatchExecgameops)
	/*
		逻辑服配置
	*/
	serverconfigService := gameservice.NewServerConfigService(config.Etcd, os.Getenv("GAME_CONFIG_NAMESPACE"))
	serverconfigHandler := gamehandler.NewServerConfigHandler(serverconfigService)
	gameRouter.Post("/config/create", serverconfigHandler.Handler_CreateServerConfig)
	gameRouter.Delete("/config/delete/:key", serverconfigHandler.Handler_DeleteServerConfig)
	gameRouter.Put("/config/update", serverconfigHandler.Handler_UpdateServerConfig)
	gameRouter.Get("/config/list", serverconfigHandler.Handler_ListServerConfig)
	gameRouter.Get("/config/listByKey/:key", serverconfigHandler.Handler_ListServerConfigBykey)
	return &gameRouter
}
