package route

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/config"
	"saurfang/internal/handler/dshandler"
	"saurfang/internal/service/datasrcservice"
)

func DataSourceRouter(app *fiber.App) *fiber.Router {
	dsservice := datasrcservice.NewDataSourceService(config.DB)
	dshandler := dshandler.NewSDataSourceHandler(dsservice)
	dsRouter := app.Group("/api/v1/ds")
	dsRouter.Post("/create", dshandler.Handler_CreateDS)
	dsRouter.Delete("/delete/:id", dshandler.Handler_DeleteDS)
	dsRouter.Put("/update/:id", dshandler.Handler_UpdateDS)
	dsRouter.Get("/list", dshandler.Handler_ShowDS)
	dsRouter.Get("/select", dshandler.Handler_SelectDS)
	return &dsRouter
}
