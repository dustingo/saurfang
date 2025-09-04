// Package tools 启动时自动处理相路由组信息到数据库
package tools

import (
	"math/rand"
	"saurfang/internal/config"
	"saurfang/internal/models/user"
	"time"
)

// PermissionData角色路由组权限数据结构
type PermissionData struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

// InitPermissionsItems 角色路由组权限入库
func InitPermissionsItems(data *PermissionData) {
	config.DB.Table("permissions").Where(user.Permission{
		Name:  data.Name,
		Group: data.Group,
	}).FirstOrCreate(&data)
}

// GenerateInviteCodes 生成邀请码
func GenerateInviteCodes() error {
	// 定义邀请码字符集
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	codes := make([]string, 100)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for j := 0; j < 100; j++ {
		b := make([]byte, 10)
		for i := range b {
			b[i] = charset[r.Intn(len(charset))]
		}
		codes[j] = string(b)
	}
	orm := config.DB
	tx := orm.Begin()
	for _, code := range codes {
		inviteCode := &user.InviteCodes{
			Code: code,
			Used: 0,
		}
		if err := tx.Create(inviteCode).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
