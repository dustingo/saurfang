package gamegroup

import "time"

// SaurfangGroups 游戏服务器归属组
type SaurfangGroups struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	Name        string     `gorm:"type:text;comment:名称" json:"name"`
	Description string     `gorm:"type:text;comment:描述" json:"description"`
}

// GroupMapping
type GroupMapping struct {
	Label string
	Value string
}
