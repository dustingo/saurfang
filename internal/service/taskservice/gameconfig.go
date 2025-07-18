package taskservice

import (
	"gorm.io/gorm"
	"saurfang/internal/models/task.go"
	"saurfang/internal/repository/base"
)

// ConfigDeployService
type ConfigDeployService struct {
	base.BaseGormRepository[task.SaurfangGameconfigtask]
	Ns string
}

// NewConfigDeployService
func NewConfigDeployService(db *gorm.DB) *ConfigDeployService {
	return &ConfigDeployService{
		BaseGormRepository: base.BaseGormRepository[task.SaurfangGameconfigtask]{DB: db},
	}
}
