package boardhandler

import (
	"saurfang/internal/config"

	nomadapi "github.com/hashicorp/nomad/api"
)

// NomadClusterStats Nomad集群统计
type NomadClusterStats struct {
	TotalNodes   int64 `json:"total_nodes"`
	OnlineNodes  int64 `json:"online_nodes"`
	OfflineNodes int64 `json:"offline_nodes"`
	TotalJobs    int64 `json:"total_jobs"`
	RunningJobs  int64 `json:"running_jobs"`
	StoppedJobs  int64 `json:"stopped_jobs"`
}

// getRealClusterStats 获取真实的Nomad集群统计
func getRealClusterStats() NomadClusterStats {
	var stats NomadClusterStats

	// 创建Nomad客户端
	if config.NomadCli == nil {
		return NomadClusterStats{
			TotalNodes:   0,
			OnlineNodes:  0,
			OfflineNodes: 0,
			TotalJobs:    0,
			RunningJobs:  0,
			StoppedJobs:  0,
		}
	}

	// 获取节点列表
	nodes, _, err := config.NomadCli.Nodes().List(&nomadapi.QueryOptions{})
	if err == nil {
		stats.TotalNodes = int64(len(nodes))
		for _, node := range nodes {
			if node.Status == "ready" {
				stats.OnlineNodes++
			} else {
				stats.OfflineNodes++
			}
		}
	}

	// 获取任务列表
	jobs, _, err := config.NomadCli.Jobs().List(&nomadapi.QueryOptions{})
	if err == nil {
		stats.TotalJobs = int64(len(jobs))
		for _, job := range jobs {
			if job.Status == "running" {
				stats.RunningJobs++
			} else {
				stats.StoppedJobs++
			}
		}
	}

	return stats
}
