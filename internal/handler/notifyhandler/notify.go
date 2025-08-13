package notifyhandler

import (
	"saurfang/internal/config"
	"saurfang/internal/models/notify"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type NtfyHandler struct {
	base.BaseGormRepository[notify.NotifySubscribe]
}

func NewNtfyHandler() *NtfyHandler {
	return &NtfyHandler{
		BaseGormRepository: base.BaseGormRepository[notify.NotifySubscribe]{DB: config.DB},
	}
}

// Handler_CreateNotifySubscribe 创建通知订阅
func (n *NtfyHandler) Handler_CreateNotifySubscribe(c fiber.Ctx) error {
	payload := new(notify.NotifySubscribe)
	if err := c.Bind().Body(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid request body", err.Error(), nil)
	}
	payload.Status = notify.StatusActive
	if err := n.Create(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to create notify subscribe", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify subscribe created successfully", "", payload)
}

// Handler_UpdateNotifySubscribe 更新通知订阅
func (n *NtfyHandler) Handler_UpdateNotifySubscribe(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid request body", err.Error(), nil)
	}
	payload := new(notify.NotifySubscribe)
	if err := c.Bind().Body(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid request body", err.Error(), nil)
	}
	if err := n.Update(uint(id), payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to update notify subscribe", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify subscribe updated successfully", "", payload)
}

// Handler_DeleteNotifySubscribe 删除通知订阅
func (n *NtfyHandler) Handler_DeleteNotifySubscribe(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid request body", err.Error(), nil)
	}
	if err := n.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to delete notify subscribe", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify subscribe deleted successfully", "", nil)
}

// Handler_ListNotifySubscribe 列出通知订阅
func (n *NtfyHandler) Handler_ListNotifySubscribe(c fiber.Ctx) error {
	subscribes, err := n.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to list notify subscribes", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify subscribes listed successfully", "", subscribes)
}
