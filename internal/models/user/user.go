package user

import "gorm.io/gorm"

// User 用户
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null" form:"username"`
	Password string `gorm:"not null" form:"password"`
	Token    string `gorm:"text" json:"token"`
	Code     string `gorm:"unique;type:varchar(100);not null" json:"code"`
	Roles    []Role `gorm:"many2many:user_roles;"`
}

// UserInfo 用户dto
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}
