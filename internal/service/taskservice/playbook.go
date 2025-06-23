package taskservice

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"saurfang/internal/models/task.go"
	"saurfang/internal/repository/base"
)

// PlaybookService
type PlaybookService struct {
	base.BaseGormRepository[task.OpsPlaybook]
	Ns string
}

// NewPlaybookService
func NewPlaybookService(etcd *clientv3.Client, ns string) *PlaybookService {
	return &PlaybookService{
		BaseGormRepository: base.BaseGormRepository[task.OpsPlaybook]{Etcd: etcd},
		Ns:                 ns,
	}
}
