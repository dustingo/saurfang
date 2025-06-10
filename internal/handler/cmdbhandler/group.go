package cmdbhandler

import (
	"saurfang/internal/models/amis"
	"saurfang/internal/models/gamegroup"
	"saurfang/internal/service/cmdbservice"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type GroupHandler struct {
	cmdbservice.GroupsService
}

func NewGroupHandler(svc *cmdbservice.GroupsService) *GroupHandler {
	return &GroupHandler{*svc}
}

// Handler_CreateNewGroup 创建新服务器归属组
func (g *GroupHandler) Handler_CreateNewGroup(c *fiber.Ctx) error {
	var group gamegroup.SaurfangGroups
	if err := c.BodyParser(&group); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := g.Create(&group); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "create success",
	})
}

// Handler_ListGroups 列出全部组
func (g *GroupHandler) Handler_ListGroups(c *fiber.Ctx) error {
	groups, err := g.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    groups,
	})
}

// Handler_UpdateGroup 组更新
func (g *GroupHandler) Handler_UpdateGroup(c *fiber.Ctx) error {
	var group gamegroup.SaurfangGroups
	id, _ := strconv.Atoi(c.Params("id"))
	if err := c.BodyParser(&group); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "请求错误",
			"err":     err.Error(),
		})
	}
	group.ID = uint(id)
	if err := g.Update(group.ID, &group); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "update failed",
			"err":     err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "upate success",
	})
}

// Handler_DeleteGroup 删除组
func (g *GroupHandler) Handler_DeleteGroup(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := g.Delete(uint(id)); err != nil {
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

// Handler_BatchDeleteGroups 批量删除组
func (g *GroupHandler) Handler_BatchDeleteGroups(c *fiber.Ctx) error {
	originIds := strings.Split(c.Query("ids"), ",")
	ids := make([]uint, 0, len(originIds))
	for _, oid := range originIds {
		id, _ := strconv.ParseUint(oid, 10, 32)
		ids = append(ids, uint(id))
	}
	if err := g.BatchDelete(ids); err != nil {
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

// Handler_GroupIdToName 在主机的crud中将组ID替换为组名
func (g *GroupHandler) Handler_GroupIdToName(c *fiber.Ctx) error {
	var gps map[string]interface{} = make(map[string]interface{})
	groups, err := g.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "show groups failed",
			"err":     err.Error(),
		})
	}
	for _, group := range *groups {
		gps[strconv.Itoa(int(group.ID))] = group.Name
	}
	gps["*"] = "未分配"
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "id to name success",
		"data":    gps,
	})
}

// Handler_SelectGroupForHost 新增主机记录时,用于选择主机归属
func (g *GroupHandler) Handler_SelectGroupForHost(c *fiber.Ctx) error {
	var amisOption amis.AmisOptions
	var amisOptions []amis.AmisOptions
	groups, err := g.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "show groups failed",
			"err":     err.Error(),
		})
	}
	for _, group := range *groups {
		amisOption.Label = group.Name
		amisOption.Value = int(group.ID)
		amisOptions = append(amisOptions, amisOption)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": 0,
		"data": fiber.Map{
			"options": amisOptions,
		},
	})
}
