package user

import "gorm.io/gorm"

// Role 用户管理角色
type Role struct {
	gorm.Model
	Name  string `gorm:"unique;not null"`
	Users []User `gorm:"many2many:user_roles;"`
}
