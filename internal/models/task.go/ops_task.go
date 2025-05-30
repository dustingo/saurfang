package task

import "time"

// SaurfangOpstask ops自定义任务库
type SaurfangOpstasks struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	Description   string     `gorm:"type:text;comment:任务描述" json:"description"`
	Hosts         string     `gorm:"type:text;comment:服务器IP" json:"hosts"`
	Playbooks     string     `gorm:"type:text;comment:playbook key" json:"playbooks"`
	User          string     `gorm:"type:text;comment:执行用户" json:"user"`
	LastExecution *time.Time `gorm:"comment:最后执行日期" json:"last_execution,omitempty"`
}

// AsynqJobs asynq数据结构
type AsynqJobs struct {
	TaskName   string `json:"task_name"`
	TaskID     int    `json:"task_id"`
	Spec       string `json:"spec"`
	TaskType   string `json:"task_type"`
	Ntfy       bool   `json:"ntfy"`
	NtfyType   string `json:"ntfy_type,omitempty"`
	NtfyTarget string `json:"ntfy_target,omitempty"`
}

// SaurfangPublishtask 服务器端发布任务
type SaurfangPublishtasks struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceID    int    `gorm:"comment:数据源ID" json:"source_id"`
	SourceLabel string `gorm:"type:text;comment:数据源标签" json:"source_label"`
	//	Dest          string     `gorm:"type:text;comment:目的目录" json:"dest"`
	Become        int        `gorm:"comment:是否启用become" json:"become"`
	BecomeUser    string     `gorm:"type:text;comment:become用户" json:"become_user"`
	Comment       string     `gorm:"type:text;comment:备注" json:"comment"`
	LastExecution *time.Time `gorm:"comment:最后执行时间" json:"last_execution,omitempty"`
	LastUser      string     `gorm:"type:text;comment:最后执行用户" json:"last_user"`
}

// SaurfangGameconfigtasks 游戏服配置任务
type SaurfangGameconfigtasks struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `gorm:"type:text;comment:任务描述" json:"description"`
	Channels    string    `gorm:"type:text;comment:渠道" json:"channels"`
}

// TemplateData 模板数据
// 用于发布时创建临时发布任务模板
type TemplateData struct {
	Hosts []Host
}
type Host struct {
	Host       string
	Become     string
	BecomeUser string // 空字符串时模板会跳过
	Tasks      []Task
}

// Task 更新为oss的结构体
type Task struct {
	Name      string
	Prefix    string // src
	Dest      string
	AccessKey string
	SecretKey string
	EndPoint  string
	Region    string
	Bucket    string
	Path      string //路径,区分版本
	Provider  string
	Profile   string
}

// OpsPlaybook etcd存储的playbook
type OpsPlaybook struct {
	Key      string
	Playbook string
}

const GamePlaybookNamespace = "game_playbook"
const GameConfigNamespace = "game_config"
const DeployConcurrency int = 10
const PrivateKeyPath string = "/root/.ssh/id_rsa"

// TaskStatus 执行playbook任务时的任务状态统计
type TaskStatus struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error"`
}

// OpsStatus 执行游戏进程开关操作的任务统计
type OpsStatus struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	ServerID     string `json:"server_id"`
	SvcName      string `json:"svc_name"`
}

// PublishTaskParams 创建任务时传参
type PublishTaskParams struct {
	Dest       string `json:"dest"`
	Become     uint   `json:"become"`
	BecomeUser string `json:"become_user"`
	Comment    string `json:"comment"`
}
