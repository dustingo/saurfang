// anisble的常规playbook任务
package taskservice

import (
	"gorm.io/gorm"
	"saurfang/internal/models/task.go"
	"saurfang/internal/repository/base"
)

// OpsTaskService
type OpsTaskService struct {
	base.BaseGormRepository[task.SaurfangOpstask]
	Ns string
}

// NewOpsTaskService
func NewOpsTaskService(db *gorm.DB) *OpsTaskService {
	return &OpsTaskService{
		BaseGormRepository: base.BaseGormRepository[task.SaurfangOpstask]{DB: db},
	}
}
