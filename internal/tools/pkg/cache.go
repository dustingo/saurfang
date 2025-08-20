package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"saurfang/internal/config"
	"saurfang/internal/models/notify"
	"strconv"
	"strings"
	"time"
)

// WarmUpCache初始化权限缓存
func WarmUpCache() error {
	type RolePermission struct {
		ID           uint   ` json:"id"`
		PermissionID string `json:"permission_id"`
		Name         string `json:"name"`
	}
	var rps []RolePermission
	query := "SELECT r.id,  rp.permission_id, p.name  FROM roles r JOIN role_permissions rp ON r.id  = rp.role_id JOIN permissions p ON rp.permission_id = p.id"
	if err := config.DB.Raw(query).Scan(&rps).Error; err != nil {
		return err
	}
	var keys []string
	for _, rp := range rps {
		key := fmt.Sprintf("role_permission:%d", rp.ID)
		config.CahceClient.Del(context.Background(), key)
		keys = append(keys, key)
	}

	for _, rp := range rps {
		key := fmt.Sprintf("role_permission:%d", rp.ID)
		if err := config.CahceClient.SAdd(context.Background(), key, rp.Name).Err(); err != nil {
			return err
		}
	}
	// 设置key过期
	for _, key := range keys {
		if err := config.CahceClient.Expire(context.Background(), key, 24*time.Hour).Err(); err != nil {
			return err
		}
	}
	return nil
}

// WarmUpNotifyCache 初始化通知订阅缓存
func WarmUpNotifyCache() error {
	var subs []notify.NotifySubscribe
	if err := config.DB.Find(&subs).Error; err != nil {
		return err
	}
	if len(subs) == 0 {
		return nil
	}
	// 清理所有相关的索引键
	pattern := fmt.Sprintf("%s:*", notify.SubscribeKey)
	keys, err := config.CahceClient.Keys(context.Background(), pattern).Result()
	if err == nil {
		for _, key := range keys {
			config.CahceClient.Del(context.Background(), key)
		}
	}

	for _, sub := range subs {
		// 存储订阅详情 - 使用 JSON 格式避免字段冗余
		detailKey := fmt.Sprintf("%s:detail:%d", notify.SubscribeKey, sub.ID)
		subData := map[string]interface{}{
			"user_id":          sub.UserID,
			"event_type":       sub.EventType,
			"notify_config_id": sub.NotifyConfigID,
			"status":           sub.Status,
		}

		// 将数据序列化为 JSON 字符串存储
		jsonData, err := json.Marshal(subData)
		if err != nil {
			continue
		}

		if err := config.CahceClient.Set(context.Background(), detailKey, jsonData, 24*time.Hour).Err(); err != nil {
			return err
		}

		// 按用户ID建立索引 - 存储订阅ID列表
		userIndexKey := fmt.Sprintf("%s:user:%d", notify.SubscribeKey, sub.UserID)
		if err := config.CahceClient.SAdd(context.Background(), userIndexKey, sub.ID).Err(); err != nil {
			return err
		}

		// 按事件类型建立索引（EventType是逗号分隔的字符串）
		if sub.EventType != "" {
			eventTypes := strings.Split(sub.EventType, ",")
			for _, eventType := range eventTypes {
				eventType = strings.TrimSpace(eventType) // 去除空格
				if eventType != "" {
					eventIndexKey := fmt.Sprintf("%s:event:%s", notify.SubscribeKey, eventType)
					if err := config.CahceClient.SAdd(context.Background(), eventIndexKey, sub.ID).Err(); err != nil {
						return err
					}
					// 设置索引键过期时间
					config.CahceClient.Expire(context.Background(), eventIndexKey, 24*time.Hour)
				}
			}
		}

		// 按通知配置ID建立索引
		configIndexKey := fmt.Sprintf("%s:config:%d", notify.SubscribeKey, sub.NotifyConfigID)
		if err := config.CahceClient.SAdd(context.Background(), configIndexKey, sub.ID).Err(); err != nil {
			return err
		}

		// 按状态建立索引
		statusIndexKey := fmt.Sprintf("%s:status:%s", notify.SubscribeKey, sub.Status)
		if err := config.CahceClient.SAdd(context.Background(), statusIndexKey, sub.ID).Err(); err != nil {
			return err
		}

		// 设置索引键过期时间
		config.CahceClient.Expire(context.Background(), userIndexKey, 24*time.Hour)
		config.CahceClient.Expire(context.Background(), configIndexKey, 24*time.Hour)
		config.CahceClient.Expire(context.Background(), statusIndexKey, 24*time.Hour)
	}

	// 加载 NotifyConfig 数据
	var configs []notify.NotifyConfig
	if err := config.DB.Find(&configs).Error; err != nil {
		return err
	}
	if len(configs) == 0 {
		return nil
	}
	for _, cfg := range configs {
		// 存储配置详情 - 使用 JSON 格式
		configDetailKey := fmt.Sprintf("%s:detail:%d", notify.ConfigKey, cfg.ID)
		configData := map[string]interface{}{
			"name":    cfg.Name,
			"channel": cfg.Channel,
			"config":  cfg.Config,
			"status":  cfg.Status,
		}

		// 将数据序列化为 JSON 字符串存储
		jsonData, err := json.Marshal(configData)
		if err != nil {
			continue
		}

		if err := config.CahceClient.Set(context.Background(), configDetailKey, jsonData, 24*time.Hour).Err(); err != nil {
			return err
		}

		// 按名称建立索引
		nameIndexKey := fmt.Sprintf("%s:name:%s", notify.ConfigKey, cfg.Name)
		if err := config.CahceClient.SAdd(context.Background(), nameIndexKey, cfg.ID).Err(); err != nil {
			return err
		}

		// 按渠道建立索引
		channelIndexKey := fmt.Sprintf("%s:channel:%s", notify.ConfigKey, cfg.Channel)
		if err := config.CahceClient.SAdd(context.Background(), channelIndexKey, cfg.ID).Err(); err != nil {
			return err
		}

		// 按状态建立索引
		configStatusIndexKey := fmt.Sprintf("%s:status:%s", notify.ConfigKey, cfg.Status)
		if err := config.CahceClient.SAdd(context.Background(), configStatusIndexKey, cfg.ID).Err(); err != nil {
			return err
		}

		// 设置配置相关键的过期时间
		config.CahceClient.Expire(context.Background(), nameIndexKey, 24*time.Hour)
		config.CahceClient.Expire(context.Background(), channelIndexKey, 24*time.Hour)
		config.CahceClient.Expire(context.Background(), configStatusIndexKey, 24*time.Hour)
	}

	return nil
}

// LoadPermissionToRedis 加载单独的角色权限缓存
func LoadPermissionToRedis(roleid uint) error {
	type RolePermission struct {
		ID           uint   ` json:"id"`
		PermissionID string `json:"permission_id"`
		Name         string `json:"name"`
	}
	var rps []RolePermission
	query := fmt.Sprintf("SELECT r.id,  rp.permission_id, p.name  FROM roles r JOIN role_permissions rp ON r.id  = rp.role_id JOIN permissions p ON rp.permission_id = p.id WHERE r.id= %d", roleid)
	if err := config.DB.Raw(query).Scan(&rps).Error; err != nil {
		return err
	}
	key := fmt.Sprintf("role_permission:%d", roleid)
	if err := config.CahceClient.Del(context.Background(), key).Err(); err != nil {
		return err
	}
	for _, rp := range rps {
		if err := config.CahceClient.SAdd(context.Background(), key, rp.Name).Err(); err != nil {
			return err
		}
	}
	if err := config.CahceClient.Expire(context.Background(), key, 24*time.Hour).Err(); err != nil {
		return err
	}
	return nil
}

// GetNotifySubscribesByUser 根据用户ID获取通知订阅
func GetNotifySubscribesByUser(userID uint) ([]notify.NotifySubscribe, error) {
	userIndexKey := fmt.Sprintf("%s:user:%d", notify.SubscribeKey, userID)
	subIDs, err := config.CahceClient.SMembers(context.Background(), userIndexKey).Result()
	if err != nil {
		return nil, err
	}

	var subs []notify.NotifySubscribe
	for _, subID := range subIDs {
		detailKey := fmt.Sprintf("%s:detail:%s", notify.SubscribeKey, subID)
		subData, err := config.CahceClient.Get(context.Background(), detailKey).Bytes()
		if err != nil {
			continue
		}

		// 解析数据
		var sub notify.NotifySubscribe
		if err := json.Unmarshal(subData, &sub); err != nil {
			continue
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

// GetNotifySubscribesByEvent 根据事件类型获取通知订阅
func GetNotifySubscribesByEvent(eventType string) ([]notify.NotifySubscribe, error) {
	eventIndexKey := fmt.Sprintf("%s:event:%s", notify.SubscribeKey, eventType)
	subIDs, err := config.CahceClient.SMembers(context.Background(), eventIndexKey).Result()
	if err != nil {
		return nil, err
	}

	var subs []notify.NotifySubscribe
	for _, subID := range subIDs {
		detailKey := fmt.Sprintf("%s:detail:%s", notify.SubscribeKey, subID)
		subData, err := config.CahceClient.Get(context.Background(), detailKey).Bytes()
		if err != nil {
			continue
		}

		// 解析数据
		var sub notify.NotifySubscribe
		if err := json.Unmarshal(subData, &sub); err != nil {
			continue
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

// GetNotifyConfigByName 根据名称获取通知配置
func GetNotifyConfigByName(name string) (*notify.NotifyConfig, error) {
	nameIndexKey := fmt.Sprintf("%s:name:%s", notify.ConfigKey, name)
	configIDs, err := config.CahceClient.SMembers(context.Background(), nameIndexKey).Result()
	if err != nil || len(configIDs) == 0 {
		return nil, fmt.Errorf("config not found for name: %s", name)
	}

	// 取第一个匹配的配置
	configID := configIDs[0]
	configDetailKey := fmt.Sprintf("%s:detail:%s", notify.ConfigKey, configID)
	configData, err := config.CahceClient.Get(context.Background(), configDetailKey).Bytes()
	if err != nil {
		return nil, err
	}

	// 解析数据
	var cfg notify.NotifyConfig
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return nil, err
	}

	// 设置ID
	if id, err := strconv.ParseUint(configID, 10, 32); err == nil {
		cfg.ID = uint(id)
	}

	return &cfg, nil
}

// GetActiveNotifyConfigs 获取所有活跃的通知配置
func GetActiveNotifyConfigs() ([]notify.NotifyConfig, error) {
	statusIndexKey := fmt.Sprintf("%s:status:%s", notify.ConfigKey, notify.StatusActive)
	configIDs, err := config.CahceClient.SMembers(context.Background(), statusIndexKey).Result()
	if err != nil {
		return nil, err
	}

	var configs []notify.NotifyConfig
	for _, configID := range configIDs {
		configDetailKey := fmt.Sprintf("%s:detail:%s", notify.ConfigKey, configID)
		configData, err := config.CahceClient.Get(context.Background(), configDetailKey).Bytes()
		if err != nil {
			continue
		}

		// 解析数据
		var cfg notify.NotifyConfig
		if err := json.Unmarshal(configData, &cfg); err != nil {
			continue
		}

		// 设置ID
		if id, err := strconv.ParseUint(configID, 10, 32); err == nil {
			cfg.ID = uint(id)
		}

		configs = append(configs, cfg)
	}

	return configs, nil
}

// GetNotifySubscribesByMultipleEvents 根据多个事件类型获取通知订阅（支持逗号分隔）
func GetNotifySubscribesByMultipleEvents(eventTypes string) ([]notify.NotifySubscribe, error) {
	if eventTypes == "" {
		return nil, fmt.Errorf("event types cannot be empty")
	}

	// 分割事件类型
	eventTypeList := strings.Split(eventTypes, ",")
	var allSubs []notify.NotifySubscribe
	seenIDs := make(map[uint]bool) // 去重

	for _, eventType := range eventTypeList {
		eventType = strings.TrimSpace(eventType)
		if eventType == "" {
			continue
		}

		subs, err := GetNotifySubscribesByEvent(eventType)
		if err != nil {
			continue // 跳过错误的，继续处理其他事件类型
		}

		// 去重添加
		for _, sub := range subs {
			if !seenIDs[sub.ID] {
				seenIDs[sub.ID] = true
				allSubs = append(allSubs, sub)
			}
		}
	}

	return allSubs, nil
}

// GetNotifySubscribesByConfig 根据通知配置ID获取通知订阅
func GetNotifySubscribesByConfig(configID uint) ([]notify.NotifySubscribe, error) {
	configIndexKey := fmt.Sprintf("%s:config:%d", notify.SubscribeKey, configID)
	subIDs, err := config.CahceClient.SMembers(context.Background(), configIndexKey).Result()
	if err != nil {
		return nil, err
	}

	var subs []notify.NotifySubscribe
	for _, subID := range subIDs {
		detailKey := fmt.Sprintf("%s:detail:%s", notify.SubscribeKey, subID)
		subData, err := config.CahceClient.Get(context.Background(), detailKey).Bytes()
		if err != nil {
			continue
		}

		// 解析数据
		var sub notify.NotifySubscribe
		if err := json.Unmarshal(subData, &sub); err != nil {
			continue
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

// GetNotifySubscribesByStatus 根据状态获取通知订阅
func GetNotifySubscribesByStatus(status string) ([]notify.NotifySubscribe, error) {
	statusIndexKey := fmt.Sprintf("%s:status:%s", notify.SubscribeKey, status)
	subIDs, err := config.CahceClient.SMembers(context.Background(), statusIndexKey).Result()
	if err != nil {
		return nil, err
	}

	var subs []notify.NotifySubscribe
	for _, subID := range subIDs {
		detailKey := fmt.Sprintf("%s:detail:%s", notify.SubscribeKey, subID)
		subData, err := config.CahceClient.Get(context.Background(), detailKey).Bytes()
		if err != nil {
			continue
		}

		// 解析数据
		var sub notify.NotifySubscribe
		if err := json.Unmarshal(subData, &sub); err != nil {
			continue
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

// GetActiveNotifySubscribes 获取所有活跃的通知订阅
func GetActiveNotifySubscribes() ([]notify.NotifySubscribe, error) {
	return GetNotifySubscribesByStatus(notify.StatusActive)
}

// GetInactiveNotifySubscribes 获取所有非活跃的通知订阅
func GetInactiveNotifySubscribes() ([]notify.NotifySubscribe, error) {
	return GetNotifySubscribesByStatus(notify.StatusInactive)
}

// GetNotifyConfigsByChannel 根据渠道获取通知配置
func GetNotifyConfigsByChannel(channel string) ([]notify.NotifyConfig, error) {
	channelIndexKey := fmt.Sprintf("%s:channel:%s", notify.ConfigKey, channel)
	configIDs, err := config.CahceClient.SMembers(context.Background(), channelIndexKey).Result()
	if err != nil {
		return nil, err
	}

	var configs []notify.NotifyConfig
	for _, configID := range configIDs {
		configDetailKey := fmt.Sprintf("%s:detail:%s", notify.ConfigKey, configID)
		configData, err := config.CahceClient.Get(context.Background(), configDetailKey).Bytes()
		if err != nil {
			continue
		}

		// 解析数据
		var cfg notify.NotifyConfig
		if err := json.Unmarshal(configData, &cfg); err != nil {
			continue
		}

		// 设置ID
		if id, err := strconv.ParseUint(configID, 10, 32); err == nil {
			cfg.ID = uint(id)
		}

		configs = append(configs, cfg)
	}

	return configs, nil
}

// GetNotifyConfigByID 根据ID获取通知配置
func GetNotifyConfigByID(configID uint) (*notify.NotifyConfig, error) {
	configDetailKey := fmt.Sprintf("%s:detail:%d", notify.ConfigKey, configID)
	configData, err := config.CahceClient.Get(context.Background(), configDetailKey).Bytes()
	if err != nil {
		return nil, err
	}

	// 解析数据
	var cfg notify.NotifyConfig
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return nil, err
	}

	cfg.ID = configID

	return &cfg, nil
}
