package taskservice

import (
	"gorm.io/gorm"
	"saurfang/internal/models/task.go"
	"saurfang/internal/repository/base"
)

// DeployService
type DeployService struct {
	base.BaseGormRepository[task.SaurfangPublishtask]
	Ns string
}

// NewDeployService
func NewDeployService(db *gorm.DB) *DeployService {
	return &DeployService{
		BaseGormRepository: base.BaseGormRepository[task.SaurfangPublishtask]{DB: db},
	}
}
