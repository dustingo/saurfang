package pkg

import (
	"context"
	"fmt"
	"saurfang/internal/config"
	"time"
)

// WarmUpCache初始化权限缓存
func WarmUpCache() error {
	type RolePermission struct {
		ID           uint   ` json:"id"`
		PermissionID string `json:"permission_id"`
		Name         string `json:"name"`
	}
	var rps []RolePermission
	query := "SELECT r.id,  rp.permission_id, p.name  FROM roles r JOIN role_permissions rp ON r.id  = rp.role_id JOIN permissions p ON rp.permission_id = p.id"
	if err := config.DB.Raw(query).Scan(&rps).Error; err != nil {
		return err
	}
	var keys []string
	for _, rp := range rps {
		key := fmt.Sprintf("role_permission:%d", rp.ID)
		config.CahceClient.Del(context.Background(), key)
		keys = append(keys, key)
	}

	for _, rp := range rps {
		key := fmt.Sprintf("role_permission:%d", rp.ID)
		if err := config.CahceClient.SAdd(context.Background(), key, rp.Name).Err(); err != nil {
			return err
		}
	}
	// 设置key过期
	for _, key := range keys {
		if err := config.CahceClient.Expire(context.Background(), key, 24*time.Hour).Err(); err != nil {
			return err
		}
	}
	return nil
}

// LoadPermissionToRedis 加载单独的角色权限缓存
func LoadPermissionToRedis(roleid uint) error {
	type RolePermission struct {
		ID           uint   ` json:"id"`
		PermissionID string `json:"permission_id"`
		Name         string `json:"name"`
	}
	var rps []RolePermission
	query := fmt.Sprintf("SELECT r.id,  rp.permission_id, p.name  FROM roles r JOIN role_permissions rp ON r.id  = rp.role_id JOIN permissions p ON rp.permission_id = p.id WHERE r.id= %d", roleid)
	if err := config.DB.Raw(query).Scan(&rps).Error; err != nil {
		return err
	}
	key := fmt.Sprintf("role_permission:%d", roleid)
	if err := config.CahceClient.Del(context.Background(), key).Err(); err != nil {
		return err
	}
	for _, rp := range rps {
		if err := config.CahceClient.SAdd(context.Background(), key, rp.Name).Err(); err != nil {
			return err
		}
	}
	if err := config.CahceClient.Expire(context.Background(), key, 24*time.Hour).Err(); err != nil {
		return err
	}
	return nil
}
