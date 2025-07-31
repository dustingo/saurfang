package datasource

import "time"

// Datasources 发布模板
// 放弃使用git，使用oss
type Datasources struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	AccessKey string    `json:"access_key"`
	SecretKey string    `json:"secret_key"`
	EndPoint  string    `json:"end_point"`
	Region    string    `json:"region"`
	Bucket    string    `json:"bucket"`
	Path      string    `json:"path"`     //区分版本
	Provider  string    `json:"provider"` // 云厂商
	Profile   string    `json:"profile"`  // remote名
	Label     string    `json:"label"`
}
