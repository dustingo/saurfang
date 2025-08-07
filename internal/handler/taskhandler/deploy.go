package taskhandler

import (
	"saurfang/internal/models/task"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type DeployHandler struct {
	base.BaseGormRepository[task.GameDeploymentTask]
	//taskservice.DeployService
}

/*
发布分为程序发布和配置发布
*/

// Handler_CreateDeployTask 创建发布任务
func (d *DeployHandler) Handler_CreateDeployTask(c fiber.Ctx) error {
	var payload task.DeployTaskPayload
	var task task.GameDeploymentTask
	if err := c.Bind().Body(&payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "bind payload error", err.Error(), nil)
	}
	if payload.ServerID == "" {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "server_id is required", "", nil)
	}
	task.ServerId = payload.ServerID
	task.Comment = payload.Comment
	if err := d.Create(&task); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "create deploy task error", err.Error(), nil)

	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_DeleteDeployTask 删除发布任务
func (d *DeployHandler) Handler_DeleteDeployTask(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "bind id error", err.Error(), nil)
	}
	if err := d.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "delete deploy task error", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_ShowDeployTask 展示发布任务
func (d *DeployHandler) Handler_ShowDeployTask(c fiber.Ctx) error {
	tasks, err := d.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "list deploy task error", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"data": tasks,
	})
}

// Handler_ShowDeployTaskByID 展示发布任务详情
func (d *DeployHandler) Handler_ShowDeployTaskByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "bind id error", err.Error(), fiber.Map{})
	}
	task, err := d.ListByID(uint(id))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "list deploy task error", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"data": task,
	})
}

// Handler_ShowDeployPerPage 展示发布任务分页
func (d *DeployHandler) Handler_ShowDeployPerPage(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize, _ := strconv.Atoi(c.Params("pageSize", "10"))
	tasks, total, err := d.ListPerPage(page, pageSize)
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "list deploy task error", err.Error(), fiber.Map{})
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"data":  tasks,
		"total": total,
	})

}
