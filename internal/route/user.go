package route

import (
	"saurfang/internal/config"
	"saurfang/internal/handler/userhandler"
	"saurfang/internal/models/user"
	"saurfang/internal/repository/base"

	"github.com/gofiber/fiber/v3"
)

type UserRouteModule struct {
	Namespace string // 路由组
	Comment   string //说明
}

func (u *UserRouteModule) Info() (namespace string, comment string) {
	return u.Namespace, u.Comment
}
func (u *UserRouteModule) RegisterRoutesModule(r *fiber.App) {
	userRoute := r.Group(u.Namespace)
	/*
		用户
	*/
	//userService := userservice.NewUserService(config.DB)
	//userHandler := userhandler.NewUserHandler(userService)
	userHandler := userhandler.UserHandler{BaseGormRepository: base.BaseGormRepository[user.User]{DB: config.DB}}
	userRoute.Post("/role/create", userHandler.Handler_CreateRole)
	userRoute.Delete("/role/delete/:id", userHandler.Handler_DeleteRole)
	userRoute.Get("/userinfo", userHandler.Handler_ShowUserInfoByRole)
	userRoute.Get("/select", userHandler.Handler_SelectUser)
	userRoute.Get("/mapping", userHandler.Handler_UserMapping)
	userRoute.Delete("/delete", userHandler.Handler_DeleteUser)
	userRoute.Get("/list", userHandler.Handler_ListUser)
	userRoute.Get("/role/list", userHandler.Handler_ListRole)
	userRoute.Put("/role/set", userHandler.Handler_SetUserRole)
	userRoute.Get("/role/select", userHandler.Handler_SelectRole)
	userRoute.Put("/role/permission/set", userHandler.Handler_SetRolePermission)
	userRoute.Get("/role/permission/select", userHandler.Handler_PermissionGroupSelect)
}
func init() {
	RegisterRoutesModule(&UserRouteModule{Namespace: "/api/v1/user", Comment: "权限管理"})
}
