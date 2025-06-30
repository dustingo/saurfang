package tools

import (
	"saurfang/internal/config"
	"saurfang/internal/models/user"
)

func RoleOfUser(id uint) uint {
	var res user.UserRole
	if err := config.DB.Debug().Table("user_roles").Where("user_id = ?", id).First(&res).Error; err != nil {
		return 0
	}
	return res.RoleID
}
