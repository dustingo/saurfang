package gamechannel

import (
	"time"
)

// SaurfangChannel 游戏服渠道
type SaurfangChannels struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	Name        string     `gorm:"type:text;comment:名称" json:"name"`
	Description string     `gorm:"type:text;comment:描述" json:"description"`
}
