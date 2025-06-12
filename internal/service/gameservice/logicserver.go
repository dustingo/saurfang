package gameservice

import (
	"saurfang/internal/models/gameserver"
	"saurfang/internal/repository/base"

	"gorm.io/gorm"
)

// LogicServerService
type LogicServerService struct {
	base.BaseGormRepository[gameserver.SaurfangGames]
}

// NewLogicServerService
func NewLogicServerService(db *gorm.DB) *LogicServerService {
	return &LogicServerService{
		BaseGormRepository: base.BaseGormRepository[gameserver.SaurfangGames]{DB: db},
	}
}
