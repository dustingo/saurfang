package route

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/config"
	"saurfang/internal/handler/userhandler"
	"saurfang/internal/service/userservice"
)

type UserRouteModule struct {
	Namespace string // 路由组
	Comment   string //说明
}

func (u *UserRouteModule) Info() (namespace string, comment string) {
	namespace = u.Namespace
	comment = u.Comment
	return u.Namespace, u.Comment
}
func (u *UserRouteModule) RegisterRoutesModule(r *fiber.App) {
	userRoute := r.Group(u.Namespace)
	/*
		用户
	*/
	userService := userservice.NewUserService(config.DB)
	userHandler := userhandler.NewUserHandler(userService)
	userRoute.Get("/userinfo", userHandler.Handler_ShowUserInfoByRole)
	userRoute.Put("/role/set", userHandler.Handler_SetUserRole)
	userRoute.Get("/role/select", userHandler.Handler_SelectRole)
	userRoute.Put("/role/permission/set", userHandler.Handler_SetRolePermission)
	userRoute.Get("/role/permission/select", userHandler.Handler_PermissionGroupSelect)
	//userRoute.Post("/auth/register", userHandler.Handler_UserRegister)
	//userRoute.Post("/auth/login", userHandler.Handler_UserLogin)
	//userRoute.Post("/auth/logout", userHandler.Handler_UserLogout)
	//userRoute.Get("/auth/status", userHandler.Handler_LoginStatus)
}
func init() {
	RegisterRoutesModule(&UserRouteModule{Namespace: "/api/v1/user", Comment: "权限管理"})
}
