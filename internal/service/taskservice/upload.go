package taskservice

import (
	"gorm.io/gorm"
	"saurfang/internal/models/upload"
	"saurfang/internal/repository/base"
)

// UploadService
type UploadService struct {
	base.BaseGormRepository[upload.UploadRecords]
}

// NewUploadService
func NewUploadService(db *gorm.DB) *UploadService {
	return &UploadService{
		BaseGormRepository: base.BaseGormRepository[upload.UploadRecords]{DB: db},
	}
}
