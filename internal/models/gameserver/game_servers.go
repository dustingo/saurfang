package gameserver

import (
	"saurfang/internal/models/gamechannel"
	"saurfang/internal/models/gamehost"
	"time"
)

// Games 游戏逻辑服
type Games struct {
	ID        uint                  `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt *time.Time            `gorm:"index" json:"deleted_at,omitempty"`
	Name      string                `gorm:"type:text;comment:名称" json:"name"`
	ServerID  string                `gorm:"type:text;comment:服务器ID" json:"server_id"`
	Status    string                `gorm:"type:text;comment:服务器状态" json:"status"`
	ChannelID *uint                 `gorm:"comment:渠道ID" json:"channel_id,omitempty"`
	Channel   *gamechannel.Channels `gorm:"foreignKey:ChannelID" json:"channel,omitempty"` // 外键关系
	ServerDir string                `gorm:"type:text;comment:服务器端家目录" json:"server_dir"`
}

// GameHosts 逻辑服与主机关系
type GameHosts struct {
	ID     uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	GameID uint           `gorm:"primaryKey" json:"game_id"`
	HostID uint           `gorm:"primaryKey" json:"host_id"`
	Game   Games          `gorm:"foreignKey:GameID" json:"game"` // 外键关系
	Host   gamehost.Hosts `gorm:"foreignKey:HostID" json:"host"` // 外键关系
}

// GameHostsDetail 逻辑服显示的挂载的游戏服务器信息
type GameHostsDetail struct {
	ID        int    `json:"id"`
	Hostname  string `json:"hostname"`
	PrivateIP string `json:"private_ip"`
}
