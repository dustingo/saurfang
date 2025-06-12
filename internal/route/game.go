package route

import (
	"saurfang/internal/config"
	"saurfang/internal/handler/gamehandler"
	"saurfang/internal/service/gameservice"

	"github.com/gofiber/fiber/v2"
)

func GameRouter(app *fiber.App) *fiber.Router {
	channelService := gameservice.NewChannelService(config.DB)
	channelHandler := gamehandler.NewHostHandler(channelService)
	gameRouter := app.Group("/api/v1/game")
	gameRouter.Post("/channel/create", channelHandler.Handler_CreateChannel)
	gameRouter.Delete("/channel/delete/:id", channelHandler.Handler_DeleteChannel)
	gameRouter.Put("/channel/update/:id", channelHandler.Handler_UpdateChannel)
	gameRouter.Get("/channel/list", channelHandler.Handler_Listhannel)
	gameRouter.Get("/channel/navlist", channelHandler.Handler_AmisNavlist)
	gameRouter.Get("/channel/select", channelHandler.Handler_SelectChannel)
	return &gameRouter
}
