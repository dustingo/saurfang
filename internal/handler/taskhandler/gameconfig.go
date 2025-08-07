// // Package task handler 游戏业务设计配置文件发布变更
package taskhandler

// import (
// 	"saurfang/internal/models/task.go"
// 	"saurfang/internal/repository/base"
// 	"saurfang/internal/tools/pkg"
// 	"strconv"

// 	"github.com/gofiber/fiber/v3"
// )

// type ConfigDeployHandler struct {
// 	base.BaseGormRepository[task.ConfigDeployTask]
// }

// // Handler_CreateConfigDeployTask 创建游戏服配置发布任务
// func (h *ConfigDeployHandler) Handler_CreateConfigDeployTask(c fiber.Ctx) error {
// 	var task task.ConfigDeployTask
// 	if err := c.Bind().Body(&task); err != nil {
// 		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "unknown task info", err.Error(), fiber.Map{})
// 	}
// 	if err := h.Create(&task); err != nil {
// 		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "fail to create task", err.Error(), fiber.Map{})
// 	}
// 	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{})
// }

// // Handler_DeleteConfigDeployTask 删除配置发布任务
// func (h *ConfigDeployHandler) Handler_DeleteConfigDeployTask(c fiber.Ctx) error {
// 	taskid, err := strconv.Atoi(c.Params("id"))
// 	if err != nil {
// 		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid task id", err.Error(), fiber.Map{})
// 	}
// 	if err := h.Delete(uint(uint(taskid))); err != nil {
// 		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "fail to delete task", err.Error(), fiber.Map{})
// 	}
// 	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{})
// }

// // Handler_ShowConfigDeployTask 展示任务
// func (h *ConfigDeployHandler) Handler_ShowConfigDeployTask(c fiber.Ctx) error {
// 	tasks, err := h.List()
// 	if err != nil {
// 		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "fail to list tasks", err.Error(), fiber.Map{})
// 	}
// 	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
// 		"data": tasks,
// 	})
// }

// // Handler_ShowConfigDeployTaskPerPage 分页显示
// func (h *ConfigDeployHandler) Handler_ShowConfigDeployTaskPerPage(c fiber.Ctx) error {
// 	page, _ := strconv.Atoi(c.Params("page", "1"))
// 	pageSize, _ := strconv.Atoi(c.Params("pageSize", "10"))
// 	tasks, count, err := h.ListPerPage(page, pageSize)
// 	if err != nil {
// 		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "fail to list tasks", err.Error(), fiber.Map{})
// 	}
// 	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
// 		"count": count,
// 		"rows":  tasks,
// 	})
// }

// // Handler_ShowConfigDeployTasksByID 很具ID查询配置发布任务
// func (h *ConfigDeployHandler) Handler_ShowConfigDeployTasksByID(c fiber.Ctx) error {
// 	id, err := strconv.Atoi(c.Params("id"))
// 	if err != nil {
// 		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "invalid task id", err.Error(), fiber.Map{})
// 	}
// 	tasks, err := h.ListByID(uint(id))
// 	if err != nil {
// 		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "fail to list tasks", err.Error(), fiber.Map{})
// 	}
// 	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
// 		"data": tasks,
// 	})
// }
