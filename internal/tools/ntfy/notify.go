package ntfy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"saurfang/internal/config"
	"saurfang/internal/models/notify"
	"strings"
	"time"
)

// Notify 通知接口
type Notify interface {
	Send(subject, message string, cnf *notify.NotifyConfig) error
}

var notifyFactory = map[string]Notify{}

// registerNotify 注册通知器
func registerNotify(name string, notify Notify) {
	notifyFactory[name] = notify
}

// Notification 通知消息的数据结构
type Notification struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func init() {
	registerNotify(notify.ChannelDingTalk, &DingTalkNotification{})
	registerNotify(notify.ChannelEmail, &EmailNotification{})
	registerNotify(notify.ChannelLark, &LarkNotification{})
}

// PublishNotification 发布消息
func PublishNotification(eventType, taskType string, successJobs []string, failedJobs []string, successCount int, failedCount int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var notification *Notification
	var notifyMsg strings.Builder
	notifyMsg.WriteString("📢 游戏操作通知: ")
	notifyMsg.WriteString(fmt.Sprintf("ℹ️任务类型: %s", taskType))
	notifyMsg.WriteString(fmt.Sprintf("✅成功任务：%d ", successCount))
	notifyMsg.WriteString(fmt.Sprintf("❌失败任务：%d ", failedCount))
	// 如果有失败的任务，添加失败ID列表
	if failedCount > 0 && len(failedJobs) > 0 {
		notifyMsg.WriteString("❌失败ID列表: ")
		// 限制显示的失败ID数量，避免消息过长
		maxDisplay := 10
		for i, job := range failedJobs {
			if i >= maxDisplay {
				notifyMsg.WriteString(fmt.Sprintf("...等共%d个 ", len(failedJobs)))
				break
			}
			notifyMsg.WriteString(fmt.Sprintf("- %s ", job))
		}
	}
	notifyMsg.WriteString(fmt.Sprintf("🕒 操作时间：%s", time.Now().Format("2006-01-02 15:04:05")))
	notification = &Notification{
		Type:    eventType,
		Message: notifyMsg.String(),
	}
	// 序列化通知消息
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		slog.Error("Error marshalling notification message", "error", err)
		return
	}
	config.NtfyClient.Publish(ctx, notify.EventChannel, notificationJSON)
}

// StartNotifySubscriber 启动Redis消息订阅监听器
func StartNotifySubscriber() {
	go func() {
		for {
			ctx := context.Background()
			// 订阅Redis通知频道
			pubsub := config.NtfyClient.Subscribe(ctx, notify.EventChannel)
			defer pubsub.Close()

			// 监听消息
			slog.Info("Started Redis notification subscriber")
			for {
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					slog.Error("Error receiving message from Redis pubsub", "error", err)
					break
				}

				// 处理接收到的消息
				slog.Info("Received notification message", "message", msg.Payload)
				// 解析通知消息
				var notification Notification
				err = json.Unmarshal([]byte(msg.Payload), &notification)
				if err != nil {
					slog.Error("Error unmarshalling notification message", "error", err)
					continue
				}
				handleNotifyEvent(notification.Type, notification.Message)

			}

			// 如果连接断开，等待一段时间后重新连接
			slog.Info("Redis notification subscriber disconnected, reconnecting in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}()
}

// handleNotifyEvent 处理通知事件的原始逻辑
func handleNotifyEvent(eventType string, message string) {
	// 查找所有的活跃订阅记录
	subscribes, err := queryNotifySubscribes()
	if err != nil {
		slog.Error("query notify subscribes error", "error", err)

		return
	}
	for _, subscribe := range subscribes {
		// 查找每个活跃订阅记录所使用的通知渠道
		configs, err := queryNotifyConfigs(subscribe.NotifyConfigID)
		if err != nil {
			slog.Error("query notify configs error", "error", err)
			continue
		}
		// 判断收到的事件是否已订阅
		subscribeEvent := strings.Split(subscribe.EventType, ",")
		if !containsString(subscribeEvent, eventType) {
			continue
		}
		if s, ok := notifyFactory[configs.Channel]; ok {
			if err = s.Send(fmt.Sprintf("📢:%s", eventType), message, configs); err != nil {
				slog.Error("send notification error", "error", err)
				continue
			}
		}
	}
}

func containsString(subscribeEvent []string, eventType string) bool {
	for _, event := range subscribeEvent {
		if strings.TrimSpace(event) == eventType {
			return true
		}
	}
	return false
}

// queryNotifySubscribes 查询订阅记录(缓存中没有订阅记录,则查询数据库,若数据库有记录则同步写入缓存)
func queryNotifySubscribes() ([]notify.NotifySubscribe, error) {
	var subscribes []notify.NotifySubscribe
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 遍历缓存中的订阅记录
	keys, err := config.CahceClient.Keys(ctx, fmt.Sprintf("%s:detail:*", notify.SubscribeKey)).Result()
	if err != nil {
		return nil, err
	}
	// 如果缓存中没有订阅记录,则查询数据库,若数据库有记录则同步写入缓存
	if len(keys) == 0 {
		err = config.DB.Model(&notify.NotifySubscribe{}).Where("status = ?", notify.StatusActive).Find(&subscribes).Error

		if err != nil {
			return nil, err
		}
		for _, subscribe := range subscribes {
			subData := map[string]interface{}{
				"user_id":          subscribe.UserID,
				"event_type":       subscribe.EventType,
				"notify_config_id": subscribe.NotifyConfigID,
				"status":           subscribe.Status,
			}
			var jsonErr error
			jsonData, jsonErr := json.Marshal(subData)
			if jsonErr != nil {
				slog.Error("marshal notify subscribe data error", "error", jsonErr)
				continue
			}
			config.CahceClient.Set(ctx, fmt.Sprintf("%s:detail:%d", notify.SubscribeKey, subscribe.ID), jsonData, 24*time.Hour)
		}
	}
	// 再次从缓存中获取
	keys, err = config.CahceClient.Keys(ctx, fmt.Sprintf("%s:detail:*", notify.SubscribeKey)).Result()
	if err != nil {
		return nil, err
	}
	// 如果缓存中没有订阅记录,则返回
	if len(keys) == 0 {
		return nil, errors.New("no notify subscribe data found in cache or database")
	}
	// 如果缓存中有订阅记录,则查询缓存
	for _, key := range keys {
		subData, err := config.CahceClient.Get(ctx, key).Bytes()
		if err != nil {
			slog.Error("get notify subscribe data error", "error", err)
			continue
		}
		var subscribe notify.NotifySubscribe
		err = json.Unmarshal(subData, &subscribe)
		if err != nil {
			slog.Error("unmarshal notify subscribe data error", "error", err)
			continue
		}
		subscribes = append(subscribes, subscribe)
	}
	return subscribes, nil
}

// queryNotifyConfigs 查询配置记录(缓存中没有配置记录,则查询数据库,若数据库有记录则同步写入缓存)
func queryNotifyConfigs(id uint) (*notify.NotifyConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 尝试从缓存获取
	key, err := config.CahceClient.Get(ctx, fmt.Sprintf("%s:config:detail:%d", notify.ConfigKey, id)).Result()
	// 如果缓存中没有或发生错误，尝试从数据库获取
	if err != nil || key == "" {
		slog.Info("Cache miss for notify config, fetching from database", "id", id, "error", err)
		var configs notify.NotifyConfig
		if err := config.DB.Model(&notify.NotifyConfig{}).Where("id = ?", id).First(&configs).Error; err != nil {
			slog.Error("Failed to get notify config from database", "id", id, "error", err)
			return &notify.NotifyConfig{}, fmt.Errorf("failed to get notify config: %w", err)
		}

		// 将数据库结果写入缓存
		jsonData, err := json.Marshal(configs)
		if err != nil {
			slog.Error("Failed to marshal notify config", "id", id, "error", err)
			return &configs, nil // 即使序列化失败，仍然返回数据库结果
		}

		if err := config.CahceClient.Set(ctx, fmt.Sprintf("%s:config:detail:%d", notify.ConfigKey, id), jsonData, 24*time.Hour).Err(); err != nil {
			slog.Error("Failed to set notify config in cache", "id", id, "error", err)
		}

		return &configs, nil
	}

	// 从缓存中解析数据
	var cfg notify.NotifyConfig

	if err := json.Unmarshal([]byte(key), &cfg); err != nil {
		slog.Error("Failed to unmarshal notify config from cache", "id", id, "error", err)

		// 解析失败，尝试从数据库重新获取
		var dbConfig notify.NotifyConfig
		if err := config.DB.Model(&notify.NotifyConfig{}).Where("id = ?", id).First(&dbConfig).Error; err != nil {
			return &notify.NotifyConfig{}, fmt.Errorf("failed to get notify config after cache unmarshal error: %w", err)

		}
		return &dbConfig, nil
	}

	return &cfg, nil
}
