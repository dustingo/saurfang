package taskhandler

import (
	"context"
	"github.com/gofiber/fiber/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/task.go"
	"saurfang/internal/service/taskservice"
	"saurfang/internal/tools"
)

// PlaybookHandler
type PlaybookHandler struct {
	taskservice.PlaybookService
}

// NewPlaybookHandler
func NewPlaybookHandler(svc *taskservice.PlaybookService) *PlaybookHandler {
	return &PlaybookHandler{*svc}
}

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
	result, err := p.Etcd.Get(context.Background(), tools.AddNamespace(key, p.Ns))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if len(result.Kvs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": "already exists",
		})
	}
	if _, err := p.Etcd.Put(context.Background(), tools.AddNamespace(key, p.Ns), playbook); err != nil {
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
	if _, err := p.Etcd.Delete(context.Background(), k); err != nil {
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
	res, err := p.Etcd.Get(context.Background(), tools.AddNamespace(payload.Key, p.Ns))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if len(res.Kvs) > 0 {
		// 存在记录就先删除后添加
		_, err := p.Etcd.Delete(context.Background(), tools.AddNamespace(payload.Key, p.Ns))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})
		}
		_, err = p.Etcd.Put(context.Background(), tools.AddNamespace(payload.Key, p.Ns), payload.Playbook)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": err.Error(),
			})
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_ShowPlaybook 展示playbook
func (p *PlaybookHandler) Handler_ShowPlaybook(c fiber.Ctx) error {
	var playbooks []task.OpsPlaybook
	var playbook task.OpsPlaybook
	res, err := p.Etcd.Get(context.Background(), p.Ns, clientv3.WithPrefix())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	for _, kv := range res.Kvs {
		playbook.Key = tools.RemoveNamespace(string(kv.Key), p.Ns)
		playbook.Playbook = string(kv.Value)
		playbooks = append(playbooks, playbook)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    playbooks,
	})
}

// Handler_ShowPlaybookByKey 指定Key查询
func (p *PlaybookHandler) Handler_ShowPlaybookByKey(c fiber.Ctx) error {
	key := c.Params("key")
	var playbook task.OpsPlaybook
	res, err := p.Etcd.Get(context.Background(), tools.AddNamespace(key, p.Ns))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	playbook.Key = string(res.Kvs[0].Key)
	playbook.Playbook = string(res.Kvs[0].Value)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    playbook,
	})

}
func (p *PlaybookHandler) Handler_PlaybookSelect(c fiber.Ctx) error {
	var playbooks []task.OpsPlaybook
	var playbook task.OpsPlaybook
	res, err := p.Etcd.Get(context.Background(), p.Ns, clientv3.WithPrefix())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	for _, kv := range res.Kvs {
		playbook.Key = tools.RemoveNamespace(string(kv.Key), p.Ns)
		playbook.Playbook = string(kv.Value)
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
