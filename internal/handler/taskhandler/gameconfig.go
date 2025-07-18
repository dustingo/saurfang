package taskhandler

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/models/task.go"
	"saurfang/internal/service/taskservice"
	"strconv"
)

type ConfigDeployHandler struct {
	taskservice.ConfigDeployService
}

func NewConfigDeployHandler(svc *taskservice.ConfigDeployService) *ConfigDeployHandler {
	return &ConfigDeployHandler{*svc}
}

// Handler_CreateConfigDeployTask 创建游戏服配置发布任务
func (h *ConfigDeployHandler) Handler_CreateConfigDeployTask(c fiber.Ctx) error {
	var task task.SaurfangGameconfigtask
	if err := c.Bind().Body(&task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "unknown task info",
		})
	}
	if err := h.Create(&task); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "fail to create task",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_DeleteConfigDeployTask 删除配置发布任务
func (h *ConfigDeployHandler) Handler_DeleteConfigDeployTask(c fiber.Ctx) error {
	taskid, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "invalid task id",
		})
	}
	if err := h.Delete(uint(uint(taskid))); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "fail to delete task",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_ShowConfigDeployTask 展示任务
func (h *ConfigDeployHandler) Handler_ShowConfigDeployTask(c fiber.Ctx) error {
	tasks, err := h.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "fail to list tasks",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    tasks,
	})
}

// Handler_ShowConfigDeployTaskPerPage 分页显示
func (h *ConfigDeployHandler) Handler_ShowConfigDeployTaskPerPage(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize, _ := strconv.Atoi(c.Params("pageSize", "10"))
	tasks, count, err := h.ListPerPage(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "fail to list tasks",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"count": count,
			"rows":  tasks,
		},
	})
}

// Handler_ShowConfigDeployTasksByID 很具ID查询配置发布任务
func (h *ConfigDeployHandler) Handler_ShowConfigDeployTasksByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "invalid task id",
		})
	}
	tasks, err := h.ListByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "fail to list tasks",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    tasks,
	})
}
