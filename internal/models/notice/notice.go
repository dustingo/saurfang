package notice

import "time"

// Notice 计划任务执行通知
type Notice struct {
	Api      string `json:"api"`
	User     string `json:"user"`
	Password string `json:"password"`
}
type PushPayload struct {
	Status      string     `json:"status"`
	AlType      string     `json:"altype"`
	Team        string     `json:"team"`
	App         string     `json:"app,omitempty"`
	Cluster     string     `json:"cluster,omitempty"`
	Name        string     `json:"name,omitempty"`
	Instance    string     `json:"instance,omitempty"`
	Description string     `json:"description"`
	Summary     string     `json:"summary,omitempty"`
	StartsAt    *time.Time `json:"startsAt,omitempty"`
	EndsAt      *time.Time `json:"endsAt,omitempty"`
	Call        int        `json:"call,omitempty"`       // 是否语音通知报警的标志
	Allocation  []string   `json:"allocation,omitempty"` // 常规情况下如果call=1是会打给team组内的成员，当指定分配的时候，就只打给allocation下的成员
}
