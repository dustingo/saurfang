package taskhandler

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/task.go"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools"
)

// PlaybookHandler
type PlaybookHandler struct {
	base.NomadJobRepository
	//taskservice.PlaybookService
}

// NewPlaybookHandler
//func NewPlaybookHandler(svc *taskservice.PlaybookService) *PlaybookHandler {
//	return &PlaybookHandler{*svc}
//}

// Handler_CreatePlaybook 创建playbook
func (p *PlaybookHandler) Handler_CreatePlaybook(c fiber.Ctx) error {
	var payload task.OpsPlaybook
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	key := payload.Key
	playbook := payload.Playbook
	err := p.CreateNomadJob(tools.AddNamespace(key, p.Ns), playbook)
	if err != nil {
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

// Handler_DeletePlaybook 删除playbook
func (p *PlaybookHandler) Handler_DeletePlaybook(c fiber.Ctx) error {
	k := c.Params("key")
	if err := p.DeleteNomadJob(tools.AddNamespace(k, p.Ns)); err != nil {
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

// Handler_UpdatePlaybook 更新playbook
func (p *PlaybookHandler) Handler_UpdatePlaybook(c fiber.Ctx) error {
	var payload task.OpsPlaybook
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := p.UpdateNomadJob(tools.AddNamespace(payload.Key, p.Ns), payload.Playbook); err != nil {
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

// Handler_ShowPlaybook 展示playbook
func (p *PlaybookHandler) Handler_ShowPlaybook(c fiber.Ctx) error {
	data, err := p.ShowNomadJob()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    data,
	})
}

// Handler_ShowPlaybookByKey 指定Key查询
func (p *PlaybookHandler) Handler_ShowPlaybookByKey(c fiber.Ctx) error {
	key := c.Params("key")
	data, err := p.ShowNomadJobByKey(tools.AddNamespace(key, p.Ns))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    data,
	})

}
func (p *PlaybookHandler) Handler_PlaybookSelect(c fiber.Ctx) error {
	var playbooks []task.OpsPlaybook
	var playbook task.OpsPlaybook
	data, err := p.ShowNomadJob()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	for _, kv := range *data {
		playbook.Key = tools.RemoveNamespace(kv.Key, p.Ns)
		playbook.Playbook = kv.Setting
		playbooks = append(playbooks, playbook)
	}
	var op amis.AmisOptionsString
	var ops []amis.AmisOptionsString
	for _, sn := range playbooks {
		op.Label = sn.Key
		op.SelectMode = "tree"
		op.Value = sn.Key
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
