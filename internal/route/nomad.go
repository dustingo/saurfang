package route

import (
	"os"
	"saurfang/internal/config"
	"saurfang/internal/handler/nomadhandler"

	"github.com/gofiber/fiber/v3"
)

type NomadRouteModule struct {
	Namespace string
	Comment   string
}

func (n *NomadRouteModule) Info() (namespace string, comment string) {
	return n.Namespace, n.Comment
}
func (n *NomadRouteModule) RegisterRoutesModule(r *fiber.App) {
	opshandler := nomadhandler.NewNomadHandler(config.ConsulCli, os.Getenv("GAME_NOMAD_JOB_NAMESPACE"))
	nomadRouter := r.Group(n.Namespace)
	nomadRouter.Get("/nodes", opshandler.Handler_ListNomadNodes)
	nomadRouter.Get("/nodes/select", opshandler.Handler_ListNomadNodesForSelect)
	nomadRouter.Get("/jobs", opshandler.Handler_ShowNomadJobs)
	nomadRouter.Post("/job/:job_id/scale", opshandler.Handler_ScaleTaskGroup)
	nomadRouter.Get("/job/:job_id/group", opshandler.Handler_ShowGroupSelect)
	nomadRouter.Delete("/job/stop", opshandler.Handler_DeployNomadOpsJob)
	nomadRouter.Put("/job/start", opshandler.Handler_DeployNomadOpsJob)
	nomadRouter.Delete("/job/purge", opshandler.Handler_PurgeNomadJob)
	handler := nomadhandler.NewNomadHandler(config.ConsulCli, os.Getenv("GAME_NOMAD_DEPLOY_NAMESPACE"))
	nomadRouter.Put("/job/deploy", handler.Handler_DeployNomadJob)
}

func init() {
	RegisterRoutesModule(&NomadRouteModule{Namespace: "/api/v1/nomad", Comment: "业务管理"})
}
