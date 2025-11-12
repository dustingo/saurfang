package gamehandler

import (
	"saurfang/internal/config"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools"
	"saurfang/internal/tools/pkg"
	"strings"

	"github.com/gofiber/fiber/v3"
	consulapi "github.com/hashicorp/consul/api"
)

// ServerConfigHandler
type ServerConfigHandler struct {
	base.NomadJobRepository
}

func NewServerConfigHandler(cli *consulapi.Client, ns string) *ServerConfigHandler {
	return &ServerConfigHandler{
		base.NomadJobRepository{Consul: cli, Nomad: config.NomadCli, Ns: ns},
	}
}

// Handler_CreateServerConfig 创建逻辑服配置
func (s *ServerConfigHandler) Handler_CreateServerConfig(c fiber.Ctx) error {
	var gcdto serverconfig.GameConfig
	if err := c.Bind().Body(&gcdto); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if err := s.CreateNomadJob(gcdto.Key, gcdto.Setting); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create server config", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_DeleteServerConfig 删除逻辑服配置
func (s *ServerConfigHandler) Handler_DeleteServerConfig(c fiber.Ctx) error {
	key := c.Query("key")
	if err := s.DeleteNomadJob(key); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete server config", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_UpdateServerConfig 更新逻辑服配置
func (s *ServerConfigHandler) Handler_UpdateServerConfig(c fiber.Ctx) error {
	var payload serverconfig.GameConfig
	if err := c.Bind().Body(&payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if err := s.UpdateNomadJob(payload.Key, payload.Setting); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to update server config", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_ListServerConfig 展示配置
func (s *ServerConfigHandler) Handler_ListServerConfig(c fiber.Ctx) error {
	data, err := s.ShowNomadJob()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list server config", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", data)
}

// Handler_ListNomadJobByKey 根据Key查询nomad job
func (s *ServerConfigHandler) Handler_ListNomadJobByKey(c fiber.Ctx) error {
	var res serverconfig.GameConfig
	key := tools.AddNamespace(c.Params("server_id"), s.Ns)
	kv := s.Consul.KV()
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list server config", err.Error(), fiber.Map{})
	}
	if pair == nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "key not found", "", fiber.Map{})
	}
	res.Key = pair.Key
	res.Setting = strings.ReplaceAll(string(pair.Value), "\r", "")
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", res)
}

// Handler_CreateNomadJob 在consul创建Nomadjob
func (s *ServerConfigHandler) Handler_CreateNomadJob(c fiber.Ctx) error {
	var payload serverconfig.GameConfig
	if err := c.Bind().Body(&payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "failed to create nomad job", err.Error(), fiber.Map{})
	}
	if err := s.CreateNomadJob(payload.Key, payload.Setting); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create nomad job", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}
