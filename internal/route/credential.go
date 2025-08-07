package route

import (
	"saurfang/internal/config"
	"saurfang/internal/handler/credentialhandler"
	"saurfang/internal/models/credential"
	"saurfang/internal/repository/base"

	"github.com/gofiber/fiber/v3"
)

type CredentialRouteModule struct {
	Namespace string // 路由组
	Comment   string //说明
}

func (c *CredentialRouteModule) Info() (namespace string, comment string) {
	return c.Namespace, c.Comment
}
func (c *CredentialRouteModule) RegisterRoutesModule(r *fiber.App) {
	credentialRouter := r.Group(c.Namespace)
	//credentialService := credentialservice.NewCredentialService(config.DB)
	//credentialHandler := credentialhandler.NewCredentialHandler(credentialService)
	credentialHandler := credentialhandler.CredentialHandler{BaseGormRepository: base.BaseGormRepository[credential.UserCredential]{DB: config.DB}}
	credentialRouter.Get("/create/:id", credentialHandler.Handler_CreateUserCredential)
	credentialRouter.Delete("/delete/:id", credentialHandler.Handler_DeleteUserCredential)
	credentialRouter.Get("/list", credentialHandler.Handler_ShowUserCredential)
	credentialRouter.Get("/status/set", credentialHandler.Handler_SetUserCredentialStatus)
}

func init() {
	RegisterRoutesModule(&CredentialRouteModule{Namespace: "/api/v1/credential", Comment: "凭证管理"})
}
