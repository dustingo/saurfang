package route

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/config"
	"saurfang/internal/handler/credentialhandler"
	"saurfang/internal/service/credentialservice"
)

type CredentialRouteModule struct {
	Namespace string // 路由组
	Comment   string //说明
}

func (c *CredentialRouteModule) Info() (namespace string, comment string) {
	namespace = c.Namespace
	comment = c.Comment
	return c.Namespace, c.Comment
}
func (c *CredentialRouteModule) RegisterRoutesModule(r *fiber.App) {
	credentialRouter := r.Group(c.Namespace)
	credentialService := credentialservice.NewCredentialService(config.DB)
	credentialHandler := credentialhandler.NewCredentialHandler(credentialService)
	credentialRouter.Get("/create/:id", credentialHandler.Handler_CreateUserCredential)
	credentialRouter.Delete("/delete/:id", credentialHandler.Handler_DeleteUserCredential)
	credentialRouter.Get("/list", credentialHandler.Handler_ShowUserCredential)
	credentialRouter.Get("/status/set", credentialHandler.Handler_SetUserCredentialStatus)
}

func init() {
	RegisterRoutesModule(&CredentialRouteModule{Namespace: "/api/v1/credential", Comment: "凭证管理"})
}
