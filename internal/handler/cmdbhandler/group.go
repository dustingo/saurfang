package cmdbhandler

import (
	"saurfang/internal/models/amis"
	"saurfang/internal/models/gamegroup"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type GroupHandler struct {
	base.BaseGormRepository[gamegroup.Groups]
	//cmdbservice.GroupsService
}

//func NewGroupHandler(svc *cmdbservice.GroupsService) *GroupHandler {
//	return &GroupHandler{*svc}
//}

// Handler_CreateNewGroup 创建新服务器归属组
func (g *GroupHandler) Handler_CreateNewGroup(c fiber.Ctx) error {
	var group gamegroup.Groups
	if err := c.Bind().Body(&group); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"err":     err.Error(),
			"message": "request error",
		})
	}
	if err := g.Create(&group); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"err":     err.Error(),
			"message": "failed to create new group",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"err":     "",
		"message": "create success",
	})
}

// Handler_ListGroups 列出全部组
func (g *GroupHandler) Handler_ListGroups(c fiber.Ctx) error {
	groups, err := g.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"err":     err.Error(),
			"message": "failed to list groups",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"err":     "",
		"message": "success",
		"data":    groups,
	})
}

// Handler_UpdateGroup 组更新
func (g *GroupHandler) Handler_UpdateGroup(c fiber.Ctx) error {
	var group gamegroup.Groups
	id, _ := strconv.Atoi(c.Params("id"))
	if err := c.Bind().Body(&group); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"err":     err.Error(),
			"message": "request error",
		})
	}
	group.ID = uint(id)
	if err := g.Update(group.ID, &group); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"err":     err.Error(),
			"message": "update failed",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"err":     "",
		"message": "upate success",
	})
}

// Handler_DeleteGroup 删除组
func (g *GroupHandler) Handler_DeleteGroup(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := g.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"err":     err.Error(),
			"message": "delete failed",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"err":     "",
		"message": "delete success",
	})
}

// Handler_BatchDeleteGroups 批量删除组
func (g *GroupHandler) Handler_BatchDeleteGroups(c fiber.Ctx) error {
	originIds := strings.Split(c.Query("ids"), ",")
	ids := make([]uint, 0, len(originIds))
	for _, oid := range originIds {
		id, _ := strconv.ParseUint(oid, 10, 32)
		ids = append(ids, uint(id))
	}
	if err := g.BatchDelete(ids); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"err":     err.Error(),
			"message": "delete failed",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"err":     "",
		"message": "delete success",
	})
}

// Handler_GroupIdToName 在主机的crud中将组ID替换为组名
func (g *GroupHandler) Handler_GroupIdToName(c fiber.Ctx) error {
	var gps map[string]interface{} = make(map[string]interface{})
	groups, err := g.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "show groups failed", err.Error(), fiber.Map{})
	}
	for _, group := range groups {
		gps[strconv.Itoa(int(group.ID))] = group.Name
	}
	gps["*"] = "未分配"
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "id to name success", "", gps)
}

// Handler_SelectGroupForHost 新增主机记录时,用于选择主机归属
func (g *GroupHandler) Handler_SelectGroupForHost(c fiber.Ctx) error {
	var amisOption amis.AmisOptions
	var amisOptions []amis.AmisOptions
	groups, err := g.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "show groups failed", err.Error(), fiber.Map{})
	}
	for _, group := range groups {
		amisOption.Label = group.Name
		amisOption.Value = int(group.ID)
		amisOptions = append(amisOptions, amisOption)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"options": amisOptions,
	})
}
