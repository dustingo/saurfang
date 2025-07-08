package cmdbhandler

import (
	"saurfang/internal/models/gamehost"
	"saurfang/internal/service/cmdbservice"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type HostHandler struct {
	cmdbservice.HostsService
}

func NewHostHandler(svc *cmdbservice.HostsService) *HostHandler {
	return &HostHandler{*svc}
}

// Handler_ListHosts 列出全部的服务器
func (h *HostHandler) Handler_ListHosts(c fiber.Ctx) error {
	hosts, err := h.BaseGormRepository.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "获取主机列表成功",
		"data":    hosts,
	})
}

// Handler_ListHostsPerPage 分页显示主机记录
func (h *HostHandler) Handler_ListHostsPerPage(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("perPage", "10"))
	//var hosts []gamehost.SaurfangHosts
	//var total int64
	//if err := h.DB.Model(&gamehost.SaurfangHosts{}).Count(&total).Error; err != nil {
	//	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	//		"status":  1,
	//		"message": err.Error(),
	//	})
	//}
	//if err := h.DB.Offset((page - 1) * pageSize).Limit(pageSize).Find(&hosts).Error; err != nil {
	//	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	//		"status":  1,
	//		"message": err.Error(),
	//	})
	//}
	hosts, total, err := h.ListPerPage(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"count": total,
			"rows":  hosts,
		},
	})
}

// Handler_CreateHost 创建主机记录
func (h *HostHandler) Handler_CreateHost(c fiber.Ctx) error {
	var host gamehost.SaurfangHosts
	if err := c.Bind().Body(&host); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "请求错误",
			"err":     err.Error(),
		})
	}
	if err := h.Create(&host); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "创建主机失败",
			"err":     err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "创建主机成功",
	})
}

// Handler_UpdateHost 根据主机ID更新主机记录
func (h *HostHandler) Handler_UpdateHost(c fiber.Ctx) error {
	var host gamehost.SaurfangHosts
	id, _ := strconv.Atoi(c.Params("id"))
	if err := c.Bind().Body(&host); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "请求错误",
			"err":     err.Error(),
		})
	}
	host.ID = uint(id)
	if err := h.Update(host.ID, &host); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "更新失败",
			"err":     err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "upate success",
	})
}

// Handler_ReGroup 为主机重新分配归属组
func (h *HostHandler) Handler_ReGroup(c fiber.Ctx) error {
	hsotId, _ := strconv.Atoi(c.Params("id"))
	groupId, _ := strconv.Atoi(c.Params("group_id"))
	if err := h.ChangeGroup(uint(hsotId), uint(groupId)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "regroup failed",
			"err":     err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "regroup success",
	})
}

// Handler_DeleteHost 删除主机记录
func (h *HostHandler) Handler_DeleteHost(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "delete failed",
			"err":     err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "delete success",
	})
}

// Handler_BatchDeleteHosts 批量删除主机记录
func (h *HostHandler) Handler_BatchDeleteHosts(c fiber.Ctx) error {
	originIds := strings.Split(c.Params("ids"), ",")
	ids := make([]uint, 0, len(originIds))
	for _, oid := range originIds {
		id, _ := strconv.ParseUint(oid, 10, 32)
		ids = append(ids, uint(id))
	}
	if err := h.BatchDelete(ids); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "delete failed",
			"err":     err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "delete success",
	})
}

// Handler_QuickSave 快速保存主机记录
func (h *HostHandler) Handler_QuickSave(c fiber.Ctx) error {
	var quickData gamehost.QuickSavePayload
	if err := c.Bind().Body(&quickData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "bad request",
			"err":     err.Error(),
		})
	}
	for _, row := range quickData.Rows {
		if err := h.Update(row.ID, &row); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": "quicksave failed",
				"err":     err.Error(),
			})
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "save success",
	})
}
