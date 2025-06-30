package userservice

import (
	"gorm.io/gorm"
	"saurfang/internal/models/user"
	"saurfang/internal/repository/base"
)

type UserService struct {
	base.BaseGormRepository[user.User]
}

// NewHostsRepository
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		BaseGormRepository: base.BaseGormRepository[user.User]{DB: db},
	}
}
