package cmdbservice

import (
	"gorm.io/gorm"
	"saurfang/internal/models/autosync"
	"saurfang/internal/repository/base"
)

type AutoSyncService struct {
	base.BaseGormRepository[autosync.SaurfangAutoSync]
}

// NewGroupsService
func NewAutoSyncService(db *gorm.DB) *AutoSyncService {
	return &AutoSyncService{
		BaseGormRepository: base.BaseGormRepository[autosync.SaurfangAutoSync]{DB: db},
	}
}
