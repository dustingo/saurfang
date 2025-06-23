package gamehandler

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/service/gameservice"
)

// ServerConfigHandler
type ServerConfigHandler struct {
	gameservice.ServerConfigService
}

func NewServerConfigHandler(svc *gameservice.ServerConfigService) *ServerConfigHandler {
	return &ServerConfigHandler{*svc}
}

// Handler_CreateServerConfig 创建逻辑服配置
func (s *ServerConfigHandler) Handler_CreateServerConfig(c fiber.Ctx) error {
	var gcdto serverconfig.GameConfigDto
	if err := c.Bind().Body(&gcdto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := s.ServerConfigService.Service_CreateGameConfig(gcdto.Key, gcdto.Setting); err != nil {
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
	key := c.Params("key")
	if err := s.ServerConfigService.Service_DeleteGameConfig(key); err != nil {
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
	var payload serverconfig.GameConfigDto
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := s.ServerConfigService.Service_UpdateGameConfig(payload.Key, payload.Setting); err != nil {
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
	data, err := s.ServerConfigService.Service_ListGameConfig()
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

// Handler_ListServerConfigBykey 根据key查询配置
func (s *ServerConfigHandler) Handler_ListServerConfigBykey(c fiber.Ctx) error {
	key := c.Params("key")
	data, err := s.ServerConfigService.Service_ListGameConfigBykey(key)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
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
