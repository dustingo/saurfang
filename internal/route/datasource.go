package route

import (
	"saurfang/internal/config"
	"saurfang/internal/handler/dshandler"
	"saurfang/internal/models/datasource"
	"saurfang/internal/repository/base"

	"github.com/gofiber/fiber/v3"
)

type DataSourceRoute struct {
	Namespace string
	Comment   string
}

func (d *DataSourceRoute) Info() (namespace string, comment string) {
	return d.Namespace, d.Comment
}
func (d *DataSourceRoute) RegisterRoutesModule(r *fiber.App) {
	dsRouter := r.Group(d.Namespace)
	//dsservice := datasrcservice.NewDataSourceService(config.DB)
	//dshandler := dshandler.NewSDataSourceHandler(dsservice)
	dshandler := dshandler.DataSourceHandler{BaseGormRepository: base.BaseGormRepository[datasource.Datasources]{DB: config.DB}}
	dsRouter.Post("/create", dshandler.Handler_CreateDS)
	dsRouter.Delete("/delete/:id", dshandler.Handler_DeleteDS)
	dsRouter.Put("/update/:id", dshandler.Handler_UpdateDS)
	dsRouter.Get("/list", dshandler.Handler_ShowDS)
	dsRouter.Get("/select", dshandler.Handler_SelectDS)
}
func init() {
	RegisterRoutesModule(&DataSourceRoute{Namespace: "/api/v1/ds", Comment: "数据源管理"})
}
