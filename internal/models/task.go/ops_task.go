package task

import "time"

// SaurfangOpstask ops自定义任务库
type SaurfangOpstask struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	Description   string     `gorm:"type:text;comment:任务描述" json:"description"`
	Hosts         string     `gorm:"type:text;comment:服务器IP" json:"hosts"`
	Playbooks     string     `gorm:"type:text;comment:playbook key" json:"playbooks"`
	User          string     `gorm:"type:text;comment:执行用户" json:"user"`
	LastExecution *time.Time `gorm:"comment:最后执行日期" json:"last_execution,omitempty"`
}

// AsynqJob asynq数据结构
type AsynqJob struct {
	TaskName   string `json:"task_name"`
	TaskID     int    `json:"task_id"`
	Spec       string `json:"spec"`
	TaskType   string `json:"task_type"`
	Ntfy       bool   `json:"ntfy"`
	NtfyType   string `json:"ntfy_type,omitempty"`
	NtfyTarget string `json:"ntfy_target,omitempty"`
}

// GameDeploymentTask 服务器端发布任务
type GameDeploymentTask struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	ServerId      string     `gorm:"type:longtext;comment:服务器ServerID" json:"server_id"`
	Comment       string     `gorm:"type:text;comment:备注" json:"comment"`
	LastExecution *time.Time `gorm:"comment:最后执行时间" json:"last_execution,omitempty"`
	LastUser      string     `gorm:"type:text;comment:最后执行用户" json:"last_user"`
}

// ConfigDeployTask 游戏服配置任务
type ConfigDeployTask struct {
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

// DeployTaskPayload 创建任务时传参
type DeployTaskPayload struct {
	ServerID string `json:"server_id"`
	Comment  string `json:"comment"`
}

// nomad job eval结果专用
type JobResult struct {
	EvalID      string `json:"evalId"`
	JobID       string `json:"job_id"`
	Type        string `json:"type"`
	TriggeredBy string `json:"triggered_by"`
	Status      string `json:"status"`
}
