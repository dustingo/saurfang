package base

import (
	"errors"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/tools"
	"strings"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	nomadapi "github.com/hashicorp/nomad/api"
)

// NomadRepository consul操作的基础接口
type NomadJobRepository struct {
	Consul *consulapi.Client
	Nomad  *nomadapi.Client
	Ns     string // consul key的前缀，区分不同的kv 功能
}

// NomadJobs nomad job下group数据结构
type NomadJobs struct {
	ID         string       `json:"id"`
	Type       string       `json:"type"`
	Status     string       `json:"status"`
	SubmitDate string       `json:"submit_date"`
	Allocation []Allocation `json:"allocation"`
}

// Allocation 分配信息
type Allocation struct {
	Group  string `json:"group"`
	Status string `json:"status"`
}

// JobStatus 作业状态常量
const (
	JobStatusComplete = "complete"
	JobStatusFailed   = "failed"
	JobStatusLost     = "lost"
	JobStatusQueued   = "queued"
	JobStatusRunning  = "success"
	JobStatusStarting = "starting"
	JobStatusUnknown  = "unknown"
)

// ShowNomadJobGroups 获取Nomad作业组列表
// t: 作业类型过滤，空字符串表示获取所有类型
func (n *NomadJobRepository) ShowNomadJobGroups(t string) ([]NomadJobs, error) {
	if n.Nomad == nil {
		return nil, errors.New("nomad client is not initialized")
	}
	var nomadJobs []NomadJobs
	//var nomadJob NomadJobs
	jobs, _, err := n.Nomad.Jobs().List(&nomadapi.QueryOptions{})
	if err != nil {
		return nomadJobs, err
	}
	for _, job := range jobs {
		if t != "" && job.Type != t {
			continue
		}
		nomadJob := n.convertJobToNomadJobs(job)
		nomadJobs = append(nomadJobs, nomadJob)
	}
	return nomadJobs, nil
}

// ScaleTaskGroup 调度指定的Job中的指定Group，返回调度ID和错误信息
func (n *NomadJobRepository) ScaleTaskGroup(jobID, target string, ops string) (string, error) {
	var count int
	switch ops {
	case "start":
		count = 1
	case "stop":
		count = 0
	default:
		return "", errors.New("invalid ops type")
	}
	resp, _, err := n.Nomad.Jobs().Scale(jobID, target, &count, ops, false, nil, &nomadapi.WriteOptions{})
	if err != nil {
		return "", err
	}
	// go func(){//save to redis}()
	// 如果通知的是不可信的nomad，需要慎重
	if !resp.KnownLeader {
		return resp.EvalID, errors.New("the system cannot confirm whether a Nomad leader currently exists. Please verify the cluster status in detail")
	}
	return resp.EvalID, nil
}
func (n *NomadJobRepository) ShowGroupsForSelect(jobID string) (*[]amis.AmisOptionsGeneric[string], error) {
	jobSummary, _, err := n.Nomad.Jobs().Summary(jobID, &nomadapi.QueryOptions{})
	if err != nil {
		return nil, err
	}
	var amisOption amis.AmisOptionsGeneric[string]
	var amisOptions []amis.AmisOptionsGeneric[string]
	for g := range jobSummary.Summary {
		amisOption.Label = g
		amisOption.Value = g
		amisOptions = append(amisOptions, amisOption)
	}
	return &amisOptions, err
}

func (n *NomadJobRepository) CreateNomadJob(key, data string) error {
	kv := n.Consul.KV()
	p := &consulapi.KVPair{
		Key:   tools.AddNamespace(key, n.Ns),
		Value: []byte(data),
	}
	_, err := kv.Put(p, nil)
	if err != nil {
		return err
	}
	return nil
}
func (n *NomadJobRepository) DeleteNomadJob(key string) error {
	kv := n.Consul.KV()
	_, err := kv.Delete(key, nil)
	if err != nil {
		return err
	}
	return nil
}
func (n *NomadJobRepository) UpdateNomadJob(key string, data string) error {
	kv := n.Consul.KV()
	pair := &consulapi.KVPair{
		Key:   key,
		Value: []byte(data),
	}
	_, err := kv.Put(pair, nil)

	if err != nil {
		return errors.New("update rejected. Key does not match or has been changed")
	}
	return nil
}
func (n *NomadJobRepository) ShowNomadJob() (*[]serverconfig.GameConfig, error) {
	var results []serverconfig.GameConfig
	var result serverconfig.GameConfig
	kv := n.Consul.KV()
	pairs, _, err := kv.List(n.Ns, nil)
	if err != nil {
		return nil, err
	}
	for _, pair := range pairs {
		result.Key = pair.Key
		result.Setting = strings.ReplaceAll(string(pair.Value), "\r", "")
		results = append(results, result)
	}
	return &results, nil
}
func (n *NomadJobRepository) ShowNomadJobByKey(key string) (*serverconfig.GameConfig, error) {
	kv := n.Consul.KV()
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, nil
	}
	result := &serverconfig.GameConfig{
		Key:     pair.Key,
		Setting: strings.ReplaceAll(string(pair.Value), "\r", ""),
	}
	return result, nil
}

// convertJobToNomadJobs 将Nomad作业转换为内部结构
func (n *NomadJobRepository) convertJobToNomadJobs(job *nomadapi.JobListStub) NomadJobs {
	var allocations []Allocation
	for group, state := range job.JobSummary.Summary {
		status := JobStatusUnknown
		if state.Running == 1 {
			status = JobStatusRunning
		} else if state.Failed == 1 {
			status = JobStatusFailed
		} else if state.Complete == 1 {
			status = JobStatusComplete
		} else if state.Queued == 1 {
			status = JobStatusQueued
		}
		allocations = append(allocations, Allocation{
			Group:  group,
			Status: status, //返回前端的颜色
		})
	}

	return NomadJobs{
		ID:         job.ID,
		Type:       job.Type,
		Status:     job.Status,
		SubmitDate: time.Unix(job.SubmitTime/1e9, job.SubmitTime%1e9).Local().Format("2006-01-02 15:04:05"),
		Allocation: allocations,
	}
}
