package route

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/config"
	"saurfang/internal/handler/dshandler"
	"saurfang/internal/service/datasrcservice"
)

type DataSourceRoute struct {
	Namespace string
	Comment   string
}

func (d *DataSourceRoute) Info() (namespace string, comment string) {
	namespace = d.Namespace
	comment = d.Comment
	return d.Namespace, d.Comment
}
func (d *DataSourceRoute) RegisterRoutesModule(r *fiber.App) {
	dsRouter := r.Group(d.Namespace)
	dsservice := datasrcservice.NewDataSourceService(config.DB)
	dshandler := dshandler.NewSDataSourceHandler(dsservice)
	dsRouter.Post("/create", dshandler.Handler_CreateDS)
	dsRouter.Delete("/delete/:id", dshandler.Handler_DeleteDS)
	dsRouter.Put("/update/:id", dshandler.Handler_UpdateDS)
	dsRouter.Get("/list", dshandler.Handler_ShowDS)
	dsRouter.Get("/select", dshandler.Handler_SelectDS)
}
func init() {
	RegisterRoutesModule(&DataSourceRoute{Namespace: "/api/v1/ds", Comment: "数据源管理"})
}

//func DataSourceRouter(app *fiber.App) *fiber.Router {
//	dsservice := datasrcservice.NewDataSourceService(config.DB)
//	dshandler := dshandler.NewSDataSourceHandler(dsservice)
//	dsRouter := app.Group("/api/v1/ds")
//	dsRouter.Post("/create", dshandler.Handler_CreateDS)
//	dsRouter.Delete("/delete/:id", dshandler.Handler_DeleteDS)
//	dsRouter.Put("/update/:id", dshandler.Handler_UpdateDS)
//	dsRouter.Get("/list", dshandler.Handler_ShowDS)
//	dsRouter.Get("/select", dshandler.Handler_SelectDS)
//	return &dsRouter
//}
