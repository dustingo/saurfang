// Package tools 启动时自动处理相路由组信息到数据库
package tools

import (
	"math/rand"
	"saurfang/internal/config"
	"saurfang/internal/models/user"
	"sync"
	"time"
)

type PermissionData struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

func InitPermissionsItems(data *PermissionData) {
	config.DB.Table("permissions").Where(user.Permission{
		Name:  data.Name,
		Group: data.Group,
	}).FirstOrCreate(&data)
}

// PermissionCache 账号权限缓存
type PermissionCache struct {
	cache map[uint]map[string]bool // 用户ID -> (权限字符串 -> bool)
	mutex sync.RWMutex             // 读写锁
	ttl   time.Duration            // 缓存有效期
}

func GenerateInviteCodes(charset string) error {
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
