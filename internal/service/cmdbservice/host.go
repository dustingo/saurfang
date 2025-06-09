package cmdbservice

import (
	"saurfang/internal/models/gamehost"
	"saurfang/internal/repository/base"

	"gorm.io/gorm"
)

// HostsRepository
type HostsService struct {
	base.BaseGormRepository[gamehost.SaurfangHosts]
}

// NewHostsRepository
func NewHostsService(db *gorm.DB) *HostsService {
	return &HostsService{
		BaseGormRepository: base.BaseGormRepository[gamehost.SaurfangHosts]{DB: db},
	}
}
