package task

import (
	"time"

	"gorm.io/gorm"
)

// CronJobs 计划任务
type CronJobs struct {
	gorm.Model
	TaskName        string           `gorm:"type:text;comment:名称" json:"task_name"`
	TaskID          int              `gorm:"type:int;comment:任务ID" json:"task_id"`
	SaurfangOpstask *SaurfangOpstask `gorm:"foreignKey:TaskID" json:"saurfang_opstask"`
	Spec            string           `gorm:"type:text;comment:计划任务表达式" json:"spec"`
	TaskType        string           `gorm:"type:text;comment:任务类型" json:"task_type"`
	TaskStatus      int              `gorm:"type:int;default:0;comment:任务状态" json:"task_status"`
	LastExecution   *time.Time       `gorm:"comment:最后执行日期" json:"last_execution,omitempty"`
	Ntfy            bool             `gorm:"type:bool;default:false;comment:是否通知" json:"ntfy"`
	NtfyType        string           `gorm:"type:text;comment:通知类型" json:"ntfy_type,omitempty"`
	NtfyTarget      string           `gorm:"type:text;comment:通知目标" json:"ntfy_target,omitempty"`
}
