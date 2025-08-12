package notify

import (
	"saurfang/internal/models/user"
	"time"
)

// status 通知订阅状态

const (
	StatusActive   string = "0"
	StatusInactive string = "1"
	SubscribeKey   string = "notify_subscribe"
	ConfigKey      string = "notify_config"
)

// NotifySubscribe 通知订阅
type NotifySubscribe struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User          user.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	EventType     string    `gorm:"type:text;not null" json:"event_type,omitempty"`
	NotifyChannel string    `gorm:"type:text;not null" json:"notify_channel,omitempty"`
	Status        string    `gorm:"type:text;not null" json:"status,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type NotifyConfig struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"type:text;not null" json:"name,omitempty"`
	Config    string    `gorm:"type:text;not null" json:"config,omitempty"`
	Status    string    `gorm:"type:text;not null" json:"status,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
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
