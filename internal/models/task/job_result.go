package task

// JobResult nomad job eval结果专用
type JobResult struct {
	EvalID      string `json:"evalId"`
	JobID       string `json:"job_id"`
	Type        string `json:"type"`
	TriggeredBy string `json:"triggered_by"`
	Status      string `json:"status"`
} 