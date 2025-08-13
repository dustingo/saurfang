package notifyhandler

import (
	"saurfang/internal/config"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/notify"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type NtfyChannelHandler struct {
	base.BaseGormRepository[notify.NotifyConfig]
}

func NewNtfyChannelHandler() *NtfyChannelHandler {
	return &NtfyChannelHandler{
		BaseGormRepository: base.BaseGormRepository[notify.NotifyConfig]{DB: config.DB},
	}
}

// Handler_CreateNotifyChannel 创建通知渠道
func (n *NtfyChannelHandler) Handler_CreateNotifyChannel(c fiber.Ctx) error {
	payload := new(notify.NotifyConfig)
	if err := c.Bind().Body(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid request body", err.Error(), nil)
	}

	// 验证必填字段
	if payload.Name == "" {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Name is required", "", nil)
	}
	if payload.Channel == "" {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Channel is required", "", nil)
	}
	if payload.Config == "" {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Config is required", "", nil)
	}

	// 验证渠道类型是否有效
	if !payload.IsValidChannel() {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid channel type", "Supported channels: dingtalk, wechat, email, http, lark", nil)
	}

	payload.Status = notify.StatusActive
	if err := n.Create(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to create notify channel", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify channel created successfully", "", payload)
}

// Handler_UpdateNotifyChannel 更新通知渠道
func (n *NtfyChannelHandler) Handler_UpdateNotifyChannel(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid request body", err.Error(), nil)
	}
	payload := new(notify.NotifyConfig)
	if err := c.Bind().Body(payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid request body", err.Error(), nil)
	}

	// 验证必填字段
	if payload.Name == "" {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Name is required", "", nil)
	}
	if payload.Channel == "" {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Channel is required", "", nil)
	}
	if payload.Config == "" {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Config is required", "", nil)
	}

	// 验证渠道类型是否有效
	if !payload.IsValidChannel() {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid channel type", "Supported channels: dingtalk, wechat, email, http, lark", nil)
	}

	payload.ID = uint(id)
	if err := n.Update(uint(id), payload); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to update notify channel", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify channel updated successfully", "", payload)
}

// Handler_DeleteNotifyChannel 删除通知渠道
func (n *NtfyChannelHandler) Handler_DeleteNotifyChannel(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Invalid request body", err.Error(), nil)
	}
	if err := n.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to delete notify channel", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify channel deleted successfully", "", nil)
}

// Handler_ListNotifyChannel 列出通知渠道
func (n *NtfyChannelHandler) Handler_ListNotifyChannel(c fiber.Ctx) error {
	channels, err := n.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to list notify channels", err.Error(), nil)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify channels listed successfully", "", channels)
}

// Handler_ListNotifyChannelForSelect 列出通知渠道用于选择
func (n *NtfyChannelHandler) Handler_ListNotifyChannelForSelect(c fiber.Ctx) error {
	var amisOption amis.AmisOptions
	var amisOptions []amis.AmisOptions
	channels, err := n.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to list notify channels", err.Error(), nil)
	}
	for _, channel := range channels {
		amisOption.Label = channel.Name
		amisOption.Value = int(channel.ID)
		amisOptions = append(amisOptions, amisOption)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify channels listed successfully", "", fiber.Map{
		"options": amisOptions,
	})
}

// Handler_ListNotifyChannelByChannel 根据渠道类型列出通知配置
func (n *NtfyChannelHandler) Handler_ListNotifyChannelByChannel(c fiber.Ctx) error {
	channelType := c.Query("channel")
	if channelType == "" {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "Channel parameter is required", "", nil)
	}

	// 使用缓存函数获取指定渠道的配置
	configs, err := pkg.GetNotifyConfigsByChannel(channelType)
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "Failed to get notify configs by channel", err.Error(), nil)
	}

	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "Notify configs listed successfully", "", configs)
}

// Handler_ListNotifyChannelMapping 列出通知渠道映射
func (n *NtfyChannelHandler) Handler_ListNotifyChannelMapping(c fiber.Ctx) error {
	var configmapping map[string]interface{} = make(map[string]interface{})
	var configs []notify.NotifyConfig
	if err := n.DB.Table("notify_configs").Find(&configs).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "fail to load users", err.Error(), fiber.Map{})
	}
	for _, config := range configs {
		configmapping[strconv.Itoa(int(config.ID))] = config.Name
	}
	configmapping["*"] = "未知"
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", configmapping)
}
