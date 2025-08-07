package route

import (
	"saurfang/internal/config"
	"saurfang/internal/handler/userhandler"
	"saurfang/internal/models/user"
	"saurfang/internal/repository/base"

	"github.com/gofiber/fiber/v3"
)

type CommonRouteModule struct {
	Namespace string // 路由组
	Comment   string //说明
}

func (u *CommonRouteModule) Info() (namespace string, comment string) {
	return u.Namespace, u.Comment
}
func (u *CommonRouteModule) RegisterRoutesModule(r *fiber.App) {
	commonRoute := r.Group(u.Namespace)
	/*
		通用：注册、登录、登出、状态
	*/
	//userService := userservice.NewUserService(config.DB)
	//userHandler := userhandler.NewUserHandler(userService)
	userHandler := userhandler.UserHandler{BaseGormRepository: base.BaseGormRepository[user.User]{DB: config.DB}}
	commonRoute.Post("/auth/register", userHandler.Handler_UserRegister)
	commonRoute.Post("/auth/login", userHandler.Handler_UserLogin)
	commonRoute.Post("/auth/logout", userHandler.Handler_UserLogout)
	commonRoute.Get("/auth/status", userHandler.Handler_LoginStatus)
}
func init() {
	RegisterRoutesModule(&CommonRouteModule{Namespace: "/api/v1/common", Comment: "通用路由"})
}
