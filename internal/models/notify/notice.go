package notify

import (
	"saurfang/internal/models/user"
	"time"

	"gorm.io/datatypes"
)

// EventType
/*
upload
gameops
gamedeploy
customjob
cronjob
*/
const (
	EventChannel        string = "event:notification"
	EventTypeUpload     string = "upload"
	EventTypeGameOps    string = "gameops"
	EventTypeGameDeploy string = "gamedeploy"
	EventTypeCustomJob  string = "customjob"
	EventTypeCronJob    string = "cronjob"
)

// status 通知订阅状态
const (
	StatusActive   string = "0"
	StatusInactive string = "1"
	SubscribeKey   string = "notify_subscribe"
	ConfigKey      string = "notify_config"
)

// 通知渠道类型常量
const (
	ChannelDingTalk string = "dingtalk"
	ChannelWeChat   string = "wechat"
	ChannelEmail    string = "email"
	ChannelHTTP     string = "http"
	ChannelLark     string = "lark"
)

// NotifySubscribe 通知订阅
type NotifySubscribe struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user_id"`
	User           user.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	EventType      string    `gorm:"type:text;not null" json:"event_type,omitempty"`
	NotifyConfigID uint      `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"notify_config_id,omitempty"`
	Status         string    `gorm:"type:text;not null" json:"status,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// NotifyConfig 通知渠道配置 数据库模型
type NotifyConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:text;not null" json:"name,omitempty"`
	Channel   string         `gorm:"type:text;not null" json:"channel,omitempty"`
	Config    datatypes.JSON `gorm:"type:json;not null" json:"config,omitempty"`
	Status    string         `gorm:"type:text;not null" json:"status,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// NotifyChannel
/*
	dingtalk :
	Config: {
		token:  "dddd",
		secret: "xxx",
	}
	email :
	Config: {
		"to":"xxxx.com,xxx.com"
	}
	wechat :
	Config: {
		AppID:          "abcdefghi",
    	AppSecret:      "jklmnopqr",
    	Token:          "mytoken",
    	EncodingAESKey: "IGNORED-IN-SANDBOX",
		ToReceivers:    "xxxx,xxx",
	}
	http:
	Config: {
		URL:         "http://localhost:8080",
		Header:      stdhttp.Header{},
		ContentType: "text/plain",
		Method:      stdhttp.MethodPost,
	}
	lark:
	Config: {
		webHookURL: "https://open.feishu.cn/open-apis/bot/v2/hook/xxxx",
	}
*/

// IsValidChannel 检查渠道是否有效
func (nc *NotifyConfig) IsValidChannel() bool {
	validChannels := []string{
		ChannelDingTalk,
		ChannelWeChat,
		ChannelEmail,
		ChannelHTTP,
		ChannelLark,
	}

	for _, validChannel := range validChannels {
		if nc.Channel == validChannel {
			return true
		}
	}
	return false
}

// GetChannelDisplayName 获取渠道的显示名称
func (nc *NotifyConfig) GetChannelDisplayName() string {
	channelNames := map[string]string{
		ChannelDingTalk: "钉钉",
		ChannelWeChat:   "微信",
		ChannelEmail:    "邮件",
		ChannelHTTP:     "HTTP",
		ChannelLark:     "飞书",
	}

	if displayName, exists := channelNames[nc.Channel]; exists {
		return displayName
	}
	return nc.Channel
}
