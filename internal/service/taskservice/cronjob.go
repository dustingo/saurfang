package taskservice

import (
	"gorm.io/gorm"
	"saurfang/internal/models/task.go"
	"saurfang/internal/repository/base"
)

// UploadService
type CronjobService struct {
	base.BaseGormRepository[task.CronJobs]
}

// NewUploadService
func NewCronjobService(db *gorm.DB) *CronjobService {
	return &CronjobService{
		BaseGormRepository: base.BaseGormRepository[task.CronJobs]{DB: db},
	}
}
