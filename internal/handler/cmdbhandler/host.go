package cmdbhandler

import (
	"saurfang/internal/models/gamehost"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type HostHandler struct {
	base.BaseGormRepository[gamehost.Hosts]
}

// Handler_ListHosts 列出全部的服务器
func (h *HostHandler) Handler_ListHosts(c fiber.Ctx) error {
	hosts, err := h.BaseGormRepository.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list hosts", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "获取主机列表成功", "", hosts)
}

// Handler_ListHostsPerPage 分页显示主机记录
func (h *HostHandler) Handler_ListHostsPerPage(c fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("perPage", "10"))
	if err != nil {
		pageSize = 10
	}
	searchType := c.Query("search_type")
	searchValue := c.Query("search_value", "")
	// 构建基础查询
	query := h.DB.Model(&gamehost.Hosts{})

	switch searchType {
	case "hostname":
		if searchValue != "" {
			query = query.Where("hostname = ?", searchValue)
		}
	case "private_ip":
		if searchValue != "" {
			query = query.Where("private_ip = ?", searchValue)
		}
	case "public_ip":
		if searchValue != "" {
			query = query.Where("public_ip = ?", searchValue)
		}
	case "labels":
		if searchValue != "" {
			query = query.Where("labels = ?", searchValue)
		}
	case "cpu":
		if searchValue != "" {
			query = query.Where("cpu = ?", searchValue)
		}
	case "memory":
		if searchValue != "" {
			query = query.Where("memory = ?", searchValue)
		}
	case "instance_id":
		if searchValue != "" {
			query = query.Where("instance_id = ?", searchValue)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to count hosts", err.Error(), fiber.Map{})
	}
	// 获取分页数据
	var data []gamehost.Hosts
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&data).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list hosts", err.Error(), fiber.Map{})
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"data":       data,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	})
}

// Handler_CreateHost 创建主机记录
func (h *HostHandler) Handler_CreateHost(c fiber.Ctx) error {
	var host gamehost.Hosts
	if err := c.Bind().Body(&host); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if err := h.Create(&host); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create host", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "create host success", "", nil)
}

// Handler_UpdateHost 根据主机ID更新主机记录
func (h *HostHandler) Handler_UpdateHost(c fiber.Ctx) error {
	var host gamehost.Hosts
	id, _ := strconv.Atoi(c.Params("id"))
	if err := c.Bind().Body(&host); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	host.ID = uint(id)
	if err := h.Update(host.ID, &host); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "update failed", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "upate success", "", nil)
}

// Handler_ReGroup 为主机重新分配归属组
func (h *HostHandler) Handler_ReGroup(c fiber.Ctx) error {
	hsotId, _ := strconv.Atoi(c.Params("id"))
	groupId, _ := strconv.Atoi(c.Params("group_id"))
	if err := h.ChangeGroup(uint(hsotId), uint(groupId)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to regroup", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "regroup success", "", nil)
}

// Handler_DeleteHost 删除主机记录
func (h *HostHandler) Handler_DeleteHost(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete host", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "delete success", "", nil)
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
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to batch delete hosts", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "delete success", "", nil)
}

// Handler_QuickSave 快速保存主机记录
func (h *HostHandler) Handler_QuickSave(c fiber.Ctx) error {
	var quickData gamehost.QuickSavePayload
	if err := c.Bind().Body(&quickData); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	for _, row := range quickData.Rows {
		if err := h.Update(row.ID, &row); err != nil {
			return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to quicksave", err.Error(), fiber.Map{})
		}
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "save success", "", nil)
}
