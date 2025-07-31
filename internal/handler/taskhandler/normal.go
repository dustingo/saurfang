package taskhandler

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/task.go"
	"saurfang/internal/repository/base"
	"strconv"
)

type OpsTaskHandler struct {
	base.BaseGormRepository[task.SaurfangOpstask]
	//taskservice.OpsTaskService
}

//func NewOpsTaskHandler(svc *taskservice.OpsTaskService) *OpsTaskHandler {
//	return &OpsTaskHandler{*svc}
//}

func (o *OpsTaskHandler) Handler_CreateOpsNormalTask(c fiber.Ctx) error {
	var task task.SaurfangOpstask
	if err := c.Bind().Body(&task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := o.Create(&task); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}
func (o *OpsTaskHandler) Handler_DeleteOpsNormalTask(c fiber.Ctx) error {
	taskId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := o.Delete(uint(taskId)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}
func (o *OpsTaskHandler) Handler_ShowOpsNormalTaskPerPage(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("perPage", "10"))
	tasks, cout, err := o.ListPerPage(page, pageSize)
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
			"count": cout,
			"rows":  tasks,
		},
	})
}
func (o *OpsTaskHandler) Handler_CrontabJobTaskSelect(c fiber.Ctx) error {
	tasks, err := o.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	var op amis.AmisOptions
	var ops []amis.AmisOptions
	for _, sn := range *tasks {
		op.Label = sn.Description
		op.Value = int(sn.ID)
		ops = append(ops, op)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"options": ops,
		},
	})
}
