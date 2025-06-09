package cmdbservice

import (
	"saurfang/internal/models/gamegroup"
	"saurfang/internal/repository/base"

	"gorm.io/gorm"
)

// GroupsService
type GroupsService struct {
	base.BaseGormRepository[gamegroup.SaurfangGroups]
}

// NewGroupsService
func NewGroupsService(db *gorm.DB) *GroupsService {
	return &GroupsService{
		BaseGormRepository: base.BaseGormRepository[gamegroup.SaurfangGroups]{DB: db},
	}
}
