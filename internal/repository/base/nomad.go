package base

import (
	"errors"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	nomadapi "github.com/hashicorp/nomad/api"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/tools"
	"strings"
	"time"
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
type Allocation struct {
	Group  string `json:"group"`
	Status string `json:"status"`
}

func (n *NomadJobRepository) ShowNomadJobGroups(t string) ([]NomadJobs, error) {
	var nomadJobs []NomadJobs
	var nomadJob NomadJobs
	jobs, _, err := n.Nomad.Jobs().List(&nomadapi.QueryOptions{})
	if err != nil {
		return nomadJobs, err
	}

	switch t {
	case "":
		for _, job := range jobs {
			var allocations []Allocation

			for group, state := range job.JobSummary.Summary {
				var status string
				if state.Running == 1 {
					status = "success"
				} else {
					status = "warning"
				}
				allocations = append(allocations, Allocation{
					Group:  group,
					Status: status,
				})
			}
			nomadJob.ID = job.ID
			nomadJob.Type = job.Type
			nomadJob.Status = job.Status
			nomadJob.SubmitDate = time.Unix(job.SubmitTime/1e9, job.SubmitTime%1e9).Local().Format("2006-01-02 15:04:05")
			nomadJob.Allocation = allocations
			nomadJobs = append(nomadJobs, nomadJob)
		}
	default:
		for _, job := range jobs {
			if job.Type != t {
				continue
			}
			var allocations []Allocation
			for group, state := range job.JobSummary.Summary {
				var status string
				if state.Running == 1 {
					status = "success"
				} else {
					status = "warning"
				}
				allocations = append(allocations, Allocation{
					Group:  group,
					Status: status,
				})
			}
			nomadJob.ID = job.ID
			nomadJob.Type = job.Type
			nomadJob.Status = job.Status
			nomadJob.SubmitDate = time.Unix(job.SubmitTime/1e9, job.SubmitTime%1e9).Local().Format("2006-01-02 15:04:05")
			nomadJob.Allocation = allocations
			nomadJobs = append(nomadJobs, nomadJob)
		}
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
		return resp.EvalID, fmt.Errorf("The system cannot confirm whether a Nomad leader currently exists. Please verify the cluster status in detail")
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
	for g, _ := range jobSummary.Summary {
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
	var result *serverconfig.GameConfig
	result.Key = pair.Key
	result.Setting = strings.ReplaceAll(string(pair.Value), "\r", "")
	return result, nil
}
