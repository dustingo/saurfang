package notice

// Notice 计划任务执行通知
type Notice struct {
	Api      string `json:"api"`
	User     string `json:"user"`
	Password string `json:"password"`
}
