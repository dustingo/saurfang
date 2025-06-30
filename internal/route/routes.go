// Package route 注册路由
package route

import "github.com/gofiber/fiber/v3"

type Router interface {
	RegisterRoutesModule(r *fiber.App)
	Info() (namespace string, comment string)
}

var RoutesModules []Router

func RegisterRoutesModule(module Router) {
	RoutesModules = append(RoutesModules, module)
}
