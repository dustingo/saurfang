package gameservice

import (
	"saurfang/internal/models/gamechannel"
	"saurfang/internal/repository/base"

	"gorm.io/gorm"
)

// ChannelService
type ChannelService struct {
	base.BaseGormRepository[gamechannel.SaurfangChannels]
}

// NewChannelService
func NewChannelService(db *gorm.DB) *ChannelService {
	return &ChannelService{
		BaseGormRepository: base.BaseGormRepository[gamechannel.SaurfangChannels]{DB: db},
	}
}
