package datasource

import "time"

// SaurfangDatasource 发布模板
// 放弃使用git，使用oss
type SaurfangDatasources struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	AccessKey string    ` json:"access_key"`
	SecretKey string    ` json:"secret_key"`
	EndPoint  string    ` json:"end_point"`
	Region    string    ` json:"region"`
	Bucket    string    ` json:"bucket"`
	Path      string    ` json:"path"`    //区分版本
	Provider  string    `json:"provider"` // 云厂商
	Profile   string    `json:"profile"`  // remote名
	Label     string    `json:"label"`
}

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
