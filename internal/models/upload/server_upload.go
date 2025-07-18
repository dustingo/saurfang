// Package upload 用于服务器端压缩包文件信息
package upload

import "time"

// FileInfo 服务器端压缩包文件信息
type FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modifiedTime"`
}

// UploadConfig 服务器端上传目录
type UploadConfig struct {
	UploadPath string `json:"uploadPath"` // 本地上传目录
	ServerPath string `json:"serverPath"` // 本地解压后的存储目录
}

// UploadRecord 服务器端上传记录
type UploadRecord struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	GameServer string    `gorm:"text" json:"game_server"`
	DestTarget string    `gorm:"text" json:"dest_target"`
	DestPath   string    `gorm:"text" json:"dest_path"`
	UploadTime time.Time `json:"upload_time"`
}
