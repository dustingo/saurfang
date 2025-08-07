package task

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CronJobs 计划任务
type CronJobs struct {
	gorm.Model
	TaskName      string     `gorm:"type:text;comment:名称" json:"task_name"`
	Spec          string     `gorm:"type:text;comment:计划任务表达式" json:"spec"`
	TaskType      string     `gorm:"type:text;comment:任务类型" json:"task_type"`
	TaskStatus    int        `gorm:"type:int;default:0;comment:任务状态" json:"task_status"`
	LastExecution *time.Time `gorm:"comment:最后执行日期" json:"last_execution,omitempty"`

	// 支持 CustomTask 和游戏服务器操作
	CustomTaskID    *uint       `gorm:"comment:关联的CustomTask ID" json:"custom_task_id,omitempty"`
	CustomTask      *CustomTask `gorm:"foreignKey:CustomTaskID" json:"custom_task,omitempty"`
	ServerIDs       string      `gorm:"type:text;comment:游戏服务器ID列表(逗号分隔)" json:"server_ids,omitempty"`
	ServerOperation string      `gorm:"type:varchar(20);comment:服务器操作类型:start,stop,restart" json:"server_operation,omitempty"`
}

// CronJobPayload 创建计划任务的请求参数
type CronJobPayload struct {
	TaskName        string   `json:"task_name"`
	CustomTaskID    *uint    `json:"custom_task_id,omitempty"`   // CustomTask ID
	ServerIDs       []string `json:"server_ids,omitempty"`       // 游戏服务器ID列表
	ServerOperation string   `json:"server_operation,omitempty"` // 服务器操作类型
	Spec            string   `json:"spec"`                       // cron表达式
	TaskType        string   `json:"task_type"`                  // 任务类型
	TaskStatus      int      `json:"task_status"`                // 任务状态
}

// UnmarshalJSON 自定义JSON反序列化，支持server_ids的字符串和数组格式
func (c *CronJobPayload) UnmarshalJSON(data []byte) error {
	// 创建一个临时结构体来处理JSON
	type Alias CronJobPayload
	aux := &struct {
		*Alias
		ServerIDsRaw interface{} `json:"server_ids,omitempty"`
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 处理server_ids字段
	if aux.ServerIDsRaw != nil {
		switch v := aux.ServerIDsRaw.(type) {
		case string:
			// 如果是字符串，按逗号分割
			if v != "" {
				c.ServerIDs = strings.Split(v, ",")
			}
		case []interface{}:
			// 如果是数组，转换为字符串数组
			for _, item := range v {
				if str, ok := item.(string); ok {
					c.ServerIDs = append(c.ServerIDs, str)
				}
			}
		case []string:
			// 如果已经是字符串数组，直接使用
			c.ServerIDs = v
		}
	}

	return nil
}

// TaskType 常量定义
const (
	TaskTypeCustom = "custom_task" // 自定义任务
	TaskTypeServer = "server_op"   // 游戏服务器操作
)

// ServerOperation 常量定义
const (
	ServerOpStart   = "start"   // 启动服务器
	ServerOpStop    = "stop"    // 停止服务器
	ServerOpRestart = "restart" // 重启服务器
)
