// Package tools 启动时自动处理相路由组信息到数据库
package tools

import (
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
	config.DB.Debug().Table("permissions").Where(user.Permission{
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

// map[userid]map[group]bool
//func NewPermissionCache(ttl time.Duration) *PermissionCache {
//	return &PermissionCache{
//		cache: make(map[uint]map[string]bool),
//		ttl:   ttl,
//	}
//}

// 从数据库加载权限
//func (pc *PermissionCache) Load(userID uint) (map[string]bool, error) {
//	var u user.UserInfo
//	// 获取用户role
//	if err := config.DB.Debug().Table("user_roles").Where("user_id = ?", userID).First(&u).Error; err != nil {
//
//	}
//	if err := config.DB.Debug().Table("users").Preload("Roles.Permissions").First(&u, userID).Error; err != nil {
//		return nil, err
//	}
//
//	permMap := make(map[string]bool)
//	for _, role := range u.Roles {
//		for _, perm := range role.Permissions {
//			key := fmt.Sprintf("%s:%s", perm.Name, perm.Group)
//			permMap[key] = true
//		}
//	}
//
//	pc.mutex.Lock()
//	pc.cache[userID] = permMap
//	pc.mutex.Unlock()
//
//	// 设置定时过期（生产环境建议改用Redis）
//	time.AfterFunc(pc.ttl, func() {
//		pc.Delete(userID)
//	})
//
//	return permMap, nil
//}
