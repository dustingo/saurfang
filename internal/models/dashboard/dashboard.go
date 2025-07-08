package dashboard

import "time"

// SaurfangTaskdashboards 任务执行次数的dashboard
type SaurfangTaskdashboards struct {
	ID    uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Task  string `gorm:"type:varchar(100);default:null" json:"task"`
	Count int    `json:"count"`
}

// LoginRecords	 用户登录记录
type LoginRecords struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string     `gorm:"type:varchar(100);default:null" json:"username"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	ClientIP  string     `gorm:"type:varchar(100);default:null" json:"client_ip"`
}

// 资源统计
type ResourceStatistics struct {
	Channels int64 `json:"channels"`
	Hosts    int64 `json:"hosts"`
	Groups   int64 `json:"groups"`
	Games    int64 `json:"games"`
	Users    int64 `json:"users"`
}
