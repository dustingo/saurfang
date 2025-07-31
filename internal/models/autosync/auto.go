package autosync

import "time"

// AutoSync 配置云厂商自动同步ECS
type AutoSync struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Cloud     string     `json:"cloud"`
	Label     string     `gorm:"unique" json:"label"`
	Region    string     `json:"region"`
	Endpoint  string     `json:"endpoint"`
	GroupID   string     `json:"group_id"`
	AccessKey string     `json:"access_key"`
	SecretKey string     `json:"secret_key"`
}
