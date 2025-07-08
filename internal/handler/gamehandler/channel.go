package gamehandler

import (
	"fmt"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/gamechannel"
	"saurfang/internal/service/gameservice"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type ChannelHandler struct {
	gameservice.ChannelService
}

func NewHostHandler(svc *gameservice.ChannelService) *ChannelHandler {
	return &ChannelHandler{*svc}
}

// Handler_CreateChannel 创建游戏服渠道
func (s *ChannelHandler) Handler_CreateChannel(c fiber.Ctx) error {
	var channel gamechannel.SaurfangChannels
	if err := c.Bind().Body(&channel); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := s.Create(&channel); err != nil {
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

// Handler_DeleteChannel 删除逻辑服渠道
func (s *ChannelHandler) Handler_DeleteChannel(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := s.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "delete failed",
			"err":     err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "delete success",
	})
}

// Handler_UpdateChannel  更新游戏服渠道
func (s *ChannelHandler) Handler_UpdateChannel(c fiber.Ctx) error {
	var channel gamechannel.SaurfangChannels
	id, _ := strconv.Atoi(c.Params("id"))
	if err := c.Bind().Body(&channel); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	channel.ID = uint(id)
	if err := s.Update(channel.ID, &channel); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "upate success",
	})
}

// Handler_Listhannel 展示渠道
func (s *ChannelHandler) Handler_Listhannel(c fiber.Ctx) error {
	channels, err := s.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": "show channel failed",
			"err":     err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    channels,
	})
}

// Handler_AmisNavlist amis的nav列表
func (s *ChannelHandler) Handler_AmisNavlist(c fiber.Ctx) error {
	var link amis.AmisChannelsNav
	var links []amis.AmisChannelsNav

	channles, err := s.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  0,
			"message": err.Error(),
		})
	}
	for _, c := range *channles {
		link.Label = c.Name
		link.To = fmt.Sprintf("?channelId=%d", c.ID)
		links = append(links, link)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    links,
	})
}

// Handler_SelectChannel 开服时选择逻辑的渠道
func (s *ChannelHandler) Handler_SelectChannel(c fiber.Ctx) error {
	channles, err := s.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  0,
			"message": err.Error(),
		})
	}
	var amisOption amis.AmisOptions
	var amisOptions []amis.AmisOptions
	for _, ch := range *channles {
		amisOption.Label = ch.Name
		amisOption.Value = int(ch.ID)
		amisOptions = append(amisOptions, amisOption)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"options": amisOptions,
		},
	})
}
