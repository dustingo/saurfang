package route

import (
	"os"
	"saurfang/internal/models/monitor"

	"github.com/gofiber/fiber/v3"
)

type MonitorRouteModule struct {
	Namespace string
	Comment   string
}

func (m *MonitorRouteModule) Info() (namespace string, comment string) {
	return m.Namespace, m.Comment
}
func (m *MonitorRouteModule) RegisterRoutesModule(r *fiber.App) {
	monitorRouter := r.Group(m.Namespace)
	monitorRouter.Get("/iframe", func(ctx fiber.Ctx) error {
		var iframe monitor.IframeSrc
		iframe.CheckMate = os.Getenv("IFRAME_CHECK_MATE_URL")
		iframe.Asynq = os.Getenv("IFRAME_ASYNQ_URL")
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  0,
			"message": "OK",
			"data":    iframe,
		})
	})

}
func init() {
	RegisterRoutesModule(&MonitorRouteModule{Namespace: "/api/v1/monitor", Comment: "监控管理"})
}
