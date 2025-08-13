package route

import (
	"saurfang/internal/handler/notifyhandler"

	"github.com/gofiber/fiber/v3"
)

type NotifyRouteModule struct {
	Namespace string
	Comment   string
}

func (n *NotifyRouteModule) Info() (namespace string, comment string) {
	return n.Namespace, n.Comment
}
func (n *NotifyRouteModule) RegisterRoutesModule(r *fiber.App) {
	notifyHandler := notifyhandler.NewNtfyHandler()
	notifyRouter := r.Group(n.Namespace)
	notifyRouter.Get("/subscribe/list", notifyHandler.Handler_ListNotifySubscribe)
	notifyRouter.Post("/subscribe/create", notifyHandler.Handler_CreateNotifySubscribe)
	notifyRouter.Put("/subscribe/update/:id", notifyHandler.Handler_UpdateNotifySubscribe)
	notifyRouter.Delete("/subscribe/delete/:id", notifyHandler.Handler_DeleteNotifySubscribe)
	//channel
	notifyChannelHandler := notifyhandler.NewNtfyChannelHandler()
	notifyRouter.Get("/channel/list", notifyChannelHandler.Handler_ListNotifyChannel)
	notifyRouter.Post("/channel/create", notifyChannelHandler.Handler_CreateNotifyChannel)
	notifyRouter.Put("/channel/update/:id", notifyChannelHandler.Handler_UpdateNotifyChannel)
	notifyRouter.Delete("/channel/delete/:id", notifyChannelHandler.Handler_DeleteNotifyChannel)
	notifyRouter.Get("/channel/select", notifyChannelHandler.Handler_ListNotifyChannelForSelect)
	notifyRouter.Get("/channel/by-channel", notifyChannelHandler.Handler_ListNotifyChannelByChannel)
	notifyRouter.Get("/channel/mapping", notifyChannelHandler.Handler_ListNotifyChannelMapping)
}

func init() {
	RegisterRoutesModule(&NotifyRouteModule{Namespace: "/api/v1/notify", Comment: "消息通知"})
}
