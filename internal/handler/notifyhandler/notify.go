package notifyhandler

import (
	"saurfang/internal/models/notify"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type NtfyHandler struct {
	base.BaseGormRepository[notify.NotifySubscribe]
}

// Handler_CreateNotifySubscribe 创建通知订阅
func (n *NtfyHandler) Handler_CreateNotifySubscribe(c fiber.Ctx) error {
	payload := new(notify.NotifySubscribe)
	if err := c.Bind().Body(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 0, "Invalid request body", err.Error(), nil)
	}
	payload.Status = notify.StatusActive
	if err := n.Create(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 0, "Failed to create notify subscribe", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 1, "Notify subscribe created successfully", "", payload)
}

// Handler_UpdateNotifySubscribe 更新通知订阅
func (n *NtfyHandler) Handler_UpdateNotifySubscribe(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 0, "Invalid request body", err.Error(), nil)
	}
	payload := new(notify.NotifySubscribe)
	if err := c.Bind().Body(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 0, "Invalid request body", err.Error(), nil)
	}
	if err := n.Update(uint(id), payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 0, "Failed to update notify subscribe", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 1, "Notify subscribe updated successfully", "", payload)
}

// Handler_DeleteNotifySubscribe 删除通知订阅
func (n *NtfyHandler) Handler_DeleteNotifySubscribe(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 0, "Invalid request body", err.Error(), nil)
	}
	if err := n.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 0, "Failed to delete notify subscribe", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 1, "Notify subscribe deleted successfully", "", nil)
}
