package gamehandler

import (
	"fmt"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/gamechannel"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type ChannelHandler struct {
	base.BaseGormRepository[gamechannel.Channels]
}

// Handler_CreateChannel 创建游戏服渠道
func (s *ChannelHandler) Handler_CreateChannel(c fiber.Ctx) error {
	var channel gamechannel.Channels
	if err := c.Bind().Body(&channel); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if err := s.Create(&channel); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create channel", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_DeleteChannel 删除逻辑服渠道
func (s *ChannelHandler) Handler_DeleteChannel(c fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := s.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete channel", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "delete success", "", nil)
}

// Handler_UpdateChannel  更新游戏服渠道
func (s *ChannelHandler) Handler_UpdateChannel(c fiber.Ctx) error {
	var channel gamechannel.Channels
	id, _ := strconv.Atoi(c.Params("id"))
	if err := c.Bind().Body(&channel); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	channel.ID = uint(id)
	if err := s.Update(channel.ID, &channel); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to update channel", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "upate success", "", nil)
}

// Handler_Listhannel 展示渠道
func (s *ChannelHandler) Handler_Listhannel(c fiber.Ctx) error {
	channels, err := s.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list channel", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", channels)
}

// Handler_AmisNavlist amis的nav列表
func (s *ChannelHandler) Handler_AmisNavlist(c fiber.Ctx) error {
	var link amis.AmisChannelsNav
	var links []amis.AmisChannelsNav

	channles, err := s.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list channel", err.Error(), fiber.Map{})
	}
	for _, c := range channles {
		link.Label = c.Name
		link.To = fmt.Sprintf("?channelId=%d", c.ID)
		links = append(links, link)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", links)
}

// Handler_SelectChannel 开服时选择逻辑的渠道
func (s *ChannelHandler) Handler_SelectChannel(c fiber.Ctx) error {
	channles, err := s.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to list channel", err.Error(), fiber.Map{})
	}
	var amisOption amis.AmisOptions
	var amisOptions []amis.AmisOptions
	for _, ch := range channles {
		amisOption.Label = ch.Name
		amisOption.Value = int(ch.ID)
		amisOptions = append(amisOptions, amisOption)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"options": amisOptions,
	})
}
