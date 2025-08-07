package task

import "time"

// GameDeploymentTask 服务器端发布任务
type GameDeploymentTask struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	ServerId      string     `gorm:"type:longtext;comment:服务器ServerID" json:"server_id"`
	Comment       string     `gorm:"type:text;comment:备注" json:"comment"`
	LastExecution *time.Time `gorm:"comment:最后执行时间" json:"last_execution,omitempty"`
	LastUser      string     `gorm:"type:text;comment:最后执行用户" json:"last_user"`
}

// DeployTaskPayload 创建任务时传参
type DeployTaskPayload struct {
	ServerID string `json:"server_id"`
	Comment  string `json:"comment"`
}
