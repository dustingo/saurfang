package task

import (
	"encoding/json"
	"time"
)

// CustomTask 自定义任务定义
type CustomTask struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `gorm:"type:varchar(255);comment:任务名称" json:"name"`
	Description string    `gorm:"type:text;comment:任务描述" json:"description"`
	ScriptType  string    `gorm:"type:varchar(20);comment:脚本类型:bash,python,shell" json:"script_type"`
	Script      string    `gorm:"type:text;comment:脚本内容" json:"script"`
	TargetHosts string    `gorm:"type:text;comment:目标主机列表" json:"target_hosts"`
	Parameters  string    `gorm:"type:text;comment:任务参数JSON" json:"parameters"`
	//Schedule    string     `gorm:"type:varchar(100);comment:调度器" json:"schedule"`
	Status     string     `gorm:"type:varchar(20);default:'active';comment:任务状态" json:"status"`
	Timeout    int        `gorm:"default:300;comment:超时时间(秒)" json:"timeout"`
	RetryCount int        `gorm:"default:0;comment:重试次数" json:"retry_count"`
	LastRun    *time.Time `gorm:"comment:最后执行时间" json:"last_run,omitempty"`
	NextRun    *time.Time `gorm:"comment:下次执行时间" json:"next_run,omitempty"`
}

// CustomTaskExecution 自定义任务执行记录
type CustomTaskExecution struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	TaskID     uint       `gorm:"comment:关联的任务ID" json:"task_id"`
	Status     string     `gorm:"type:varchar(20);comment:执行状态:running,success,failed,pending,complete" json:"status"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Result     string     `gorm:"type:text;comment:执行结果" json:"result"`
	ErrorMsg   string     `gorm:"type:text;comment:错误信息" json:"error_msg"`
	NomadJobID string     `gorm:"type:varchar(255);comment:关联的Nomad Job ID" json:"nomad_job_id"`
	ExitCode   int        `gorm:"comment:退出码" json:"exit_code"`
	// 新增字段用于跟踪 Nomad Job 状态
	NomadEvalID    string     `gorm:"type:varchar(255);comment:Nomad Evaluation ID" json:"nomad_eval_id"`
	NomadAllocID   string     `gorm:"type:varchar(255);comment:Nomad Allocation ID" json:"nomad_alloc_id"`
	NomadNodeID    string     `gorm:"type:varchar(255);comment:Nomad Node ID" json:"nomad_node_id"`
	NomadJobStatus string     `gorm:"type:text;comment:Nomad Job状态" json:"nomad_job_status"`
	LastCheckTime  *time.Time `gorm:"comment:最后检查时间" json:"last_check_time"`
	CheckCount     int        `gorm:"default:0;comment:检查次数" json:"check_count"`
	MaxCheckCount  int        `gorm:"default:60;comment:最大检查次数" json:"max_check_count"`
	CheckInterval  int        `gorm:"default:10;comment:检查间隔(秒)" json:"check_interval"`
}

// CustomTaskPayload 创建自定义任务的请求参数
type CustomTaskPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ScriptType  string `json:"script_type"`
	Script      string `json:"script"`
	TargetHosts string `json:"target_hosts"`
	Parameters  string `json:"parameters"` // 改为 string 类型，支持 JSON 字符串
	//Schedule    string                 `json:"schedule"`
	Timeout    int `json:"timeout"`
	RetryCount int `json:"retry_count"`
}

// UnmarshalJSON 自定义 JSON 解析，支持 parameters 为对象或字符串
func (p *CustomTaskPayload) UnmarshalJSON(data []byte) error {
	// 使用临时结构体来避免递归调用
	type Alias CustomTaskPayload
	aux := &struct {
		Parameters interface{} `json:"parameters"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 处理 parameters 字段
	if aux.Parameters != nil {
		switch v := aux.Parameters.(type) {
		case string:
			// 如果已经是字符串，直接使用
			p.Parameters = v
		case map[string]interface{}:
			// 如果是对象，转换为 JSON 字符串
			if jsonBytes, err := json.Marshal(v); err == nil {
				p.Parameters = string(jsonBytes)
			}
		default:
			// 其他类型，尝试转换为 JSON 字符串
			if jsonBytes, err := json.Marshal(v); err == nil {
				p.Parameters = string(jsonBytes)
			}
		}
	}

	return nil
}

// ScriptTemplate 脚本模板
type ScriptTemplate struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `gorm:"type:varchar(255);comment:模板名称" json:"name"`
	Description string    `gorm:"type:text;comment:模板描述" json:"description"`
	Category    string    `gorm:"type:varchar(50);comment:模板分类" json:"category"`
	ScriptType  string    `gorm:"type:varchar(20);comment:脚本类型" json:"script_type"`
	Template    string    `gorm:"type:text;comment:模板内容" json:"template"`
	Parameters  string    `gorm:"type:text;comment:参数定义JSON" json:"parameters"`
	Tags        string    `gorm:"type:text;comment:标签" json:"tags"`
}

// ScriptTemplatePayload 创建脚本模板的请求参数
type ScriptTemplatePayload struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	ScriptType  string                 `json:"script_type"`
	Template    string                 `json:"template"`
	Parameters  map[string]interface{} `json:"parameters"`
	Tags        []string               `json:"tags"`
}
