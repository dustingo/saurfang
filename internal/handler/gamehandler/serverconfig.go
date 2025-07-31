package gamehandler

import (
	"github.com/gofiber/fiber/v3"
	consulapi "github.com/hashicorp/consul/api"
	"golang.org/x/exp/slog"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools"
	"strings"
)

// ServerConfigHandler
type ServerConfigHandler struct {
	base.NomadJobRepository
}

func NewServerConfigHandler(cli *consulapi.Client, ns string) *ServerConfigHandler {
	client, err := config.NewNomadClient()
	if err != nil {
		slog.Error("init nomad client err:", err)
		os.Exit(-1)
	}
	return &ServerConfigHandler{
		base.NomadJobRepository{Consul: cli, Nomad: client, Ns: ns},
	}
}

// Handler_CreateServerConfig 创建逻辑服配置
func (s *ServerConfigHandler) Handler_CreateServerConfig(c fiber.Ctx) error {
	var gcdto serverconfig.GameConfig
	if err := c.Bind().Body(&gcdto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := s.CreateNomadJob(gcdto.Key, gcdto.Setting); err != nil {
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

// Handler_DeleteServerConfig 删除逻辑服配置
func (s *ServerConfigHandler) Handler_DeleteServerConfig(c fiber.Ctx) error {
	key := c.Query("key")
	if err := s.DeleteNomadJob(key); err != nil {
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

// Handler_UpdateServerConfig 更新逻辑服配置
func (s *ServerConfigHandler) Handler_UpdateServerConfig(c fiber.Ctx) error {
	var payload serverconfig.GameConfig
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := s.UpdateNomadJob(payload.Key, payload.Setting); err != nil {
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

// Handler_ListServerConfig 展示配置
func (s *ServerConfigHandler) Handler_ListServerConfig(c fiber.Ctx) error {
	data, err := s.ShowNomadJob()
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

// Handler_ListNomadJobByKey 根据Key查询nomad job
func (s *ServerConfigHandler) Handler_ListNomadJobByKey(c fiber.Ctx) error {
	var res serverconfig.GameConfig
	key := tools.AddNamespace(c.Params("server_id"), s.Ns)
	kv := s.Consul.KV()
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if pair == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "key not found",
		})
	}
	res.Key = pair.Key
	res.Setting = strings.ReplaceAll(string(pair.Value), "\r", "")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    res,
	})
}

// Handler_CreateNomadJob 在consul创建Nomadjob
func (s *ServerConfigHandler) Handler_CreateNomadJob(c fiber.Ctx) error {
	var payload serverconfig.GameConfig
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := s.CreateNomadJob(payload.Key, payload.Setting); err != nil {
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
