package taskhandler

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/models/task.go"
	"saurfang/internal/repository/base"
	"strconv"
)

type CronjobHandler struct {
	base.BaseGormRepository[task.CronJobs]
	//taskservice.CronjobService
}

//	func NewCronjobHandler(svc *taskservice.CronjobService) *CronjobHandler {
//		return &CronjobHandler{*svc}
//	}
func (j *CronjobHandler) Handler_CreateCronjobTask(c fiber.Ctx) error {
	var synqJobs task.AsynqJob
	if err := c.Bind().Body(&synqJobs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	var cronJob task.CronJobs
	cronJob.TaskID = synqJobs.TaskID
	cronJob.TaskName = synqJobs.TaskName
	cronJob.Spec = synqJobs.Spec
	cronJob.TaskType = synqJobs.TaskType
	cronJob.Ntfy = synqJobs.Ntfy
	cronJob.TaskType = synqJobs.TaskType
	cronJob.NtfyTarget = synqJobs.NtfyTarget
	if err := j.Create(&cronJob); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}
func (j *CronjobHandler) Handler_DeleteCronjobTask(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := j.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}
func (j *CronjobHandler) Handler_UpdateCronjobTask(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var cronJob task.CronJobs
	if err := c.Bind().Body(&cronJob); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := j.Update(uint(id), &cronJob); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}
func (j *CronjobHandler) Handler_ShowCronjobTask(c fiber.Ctx) error {
	cronjobs, err := j.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    cronjobs,
	})
}
func (j *CronjobHandler) Handler_ResetCronjobStatus(c fiber.Ctx) error {
	id := c.Params("id")
	if err := j.DB.Model(&task.CronJobs{}).Where("id = ?", id).Update("task_status", 0).Error; err != nil {
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
