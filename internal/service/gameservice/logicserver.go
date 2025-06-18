package gameservice

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"saurfang/internal/models/gameserver"
	"saurfang/internal/repository/base"

	"gorm.io/gorm"
)

// LogicServerService
type LogicServerService struct {
	base.BaseGormRepository[gameserver.SaurfangGames]
}

// NewLogicServerService
func NewLogicServerService(db *gorm.DB, etcd *clientv3.Client) *LogicServerService {
	return &LogicServerService{
		BaseGormRepository: base.BaseGormRepository[gameserver.SaurfangGames]{DB: db, Etcd: etcd},
	}
}
