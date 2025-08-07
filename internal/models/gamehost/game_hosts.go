package gamehost

import (
	"saurfang/internal/models/gamegroup"
	"time"
)

// Hosts 主机
type Hosts struct {
	ID           uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	DeletedAt    *time.Time        `gorm:"index" json:"deleted_at,omitempty"`
	InstanceID   string            `gorm:"unique" json:"instance_id"`
	InstanceType string            `gorm:"type:text;comment:云服务器规格" json:"instance_type"`
	Hostname     string            `gorm:"type:text;comment:主机名" json:"hostname"`
	PublicIP     string            `gorm:"type:text;comment:公网IP" json:"public_ip"`
	PrivateIP    string            `gorm:"type:text;comment:内网IP" json:"private_ip"`
	CPU          string            `gorm:"type:text;comment:CPU" json:"cpu"`
	Memory       string            `gorm:"type:text;comment:内存" json:"memory"`
	OsName       string            `gorm:"type:text;comment:系统" json:"os_name"`
	Port         int               `gorm:"default:null" json:"port"`
	Labels       string            `gorm:"type:text;comment:标签" json:"labels"`
	GroupID      *uint             `gorm:"comment:组ID" json:"group_id,omitempty"`
	Group        *gamegroup.Groups `gorm:"foreignKey:GroupID" json:"group,omitempty"` // 外键关系
}

// QuickSavePayload 服务器快速保存
type QuickSavePayload struct {
	Rows            []Hosts     `json:"rows"`
	RowsDiff        interface{} `json:"rowsDiff"`
	Indexes         []string    `json:"indexes"`
	RowsOrigin      interface{} `json:"rowsOrigin"`
	UnModifiedItems interface{} `json:"unModifiedItems"`
}
