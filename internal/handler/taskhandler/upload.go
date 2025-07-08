package taskhandler

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"io"
	"os"
	"path"
	"path/filepath"
	"saurfang/internal/config"
	"saurfang/internal/models/upload"
	"saurfang/internal/service/taskservice"
	"saurfang/internal/tools"
	"strconv"
	"time"
)

type UploadHandler struct {
	taskservice.UploadService
}

func NewUploadHandler(svc *taskservice.UploadService) *UploadHandler {
	return &UploadHandler{*svc}
}
func (u *UploadHandler) Handler_ShowServerPackage(c fiber.Ctx) error {
	var files []upload.FileInfo
	entries, err := os.ReadDir(os.Getenv("SERVER_PACKAGE_SRC_PATH"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		files = append(files, upload.FileInfo{
			Name:         entry.Name(),
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    files,
	})
}
func (u *UploadHandler) Handler_UploadServerPackage(c fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	file := c.Query("file")
	targetID, _ := strconv.Atoi(c.Query("target"))
	startTime := time.Now()
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		if _, err := os.Stat(path.Join(os.Getenv("SERVER_PACKAGE_SRC_PATH"), file)); err != nil {
			writer.Write([]byte(fmt.Sprintf("[%v] ERROR 缺少服务器端文件\n", time.Now().Format("2006-01-02 13:04:05"))))
		}
		writer.Write([]byte(fmt.Sprintf("[%v] INFO 清空目标目录 %s\n", time.Now().Format("2006-01-02 13:04:05"), os.Getenv("SERVER_PACKAGE_SRC_PATH"))))
		files, err := filepath.Glob(path.Join(os.Getenv("SERVER_PACKAGE_DEST_PATH"), "*"))
		if err != nil {
			writer.Write([]byte(fmt.Sprintf("[%v] ERROR 清空目录失败\n", time.Now().Format("2006-01-02 13:04:05"))))
		}
		for _, f := range files {
			err = os.RemoveAll(f)
			if err != nil {
				writer.Write([]byte(fmt.Sprintf("[%v] ERROR 删除 %s 失败: %v\n", time.Now().Format("2006-01-02 13:04:05"), f, err.Error())))
			}
		}
		writer.Write([]byte(fmt.Sprintf("[%v] Success 清空目录成功\n", time.Now().Format("2006-01-02 13:04:05"))))
		writer.Write([]byte(fmt.Sprintf("[%v] INFO 正在解压服务器端 %s 到 %s\n", time.Now().Format("2006-01-02 13:04:05"), file, os.Getenv("SERVER_PACKAGE_DEST_PATH"))))
		if err := tools.SafeUnzip(path.Join(os.Getenv("SERVER_PACKAGE_SRC_PATH"), file), os.Getenv("SERVER_PACKAGE_DEST_PATH")); err != nil {
			writer.Write([]byte(fmt.Sprintf("[%v] ERROR 解压服务器端失败 %s\n", time.Now().Format("2006-01-02 13:04:05"), err.Error())))
		}
		entries, err := os.ReadDir(os.Getenv("SERVER_PACKAGE_DEST_PATH"))
		if err != nil {
			writer.Write([]byte(fmt.Sprintf("[%v] ERROR 获取资源列表失败 %s\n", time.Now().Format("2006-01-02 13:04:05"), err.Error())))

		}
		for _, entry := range entries {
			info, _ := entry.Info()
			if entry.IsDir() {
				writer.Write([]byte(fmt.Sprintf("%s %s %d %s\n", info.Mode().String(), info.ModTime().String(), info.Size(), entry.Name())))

			} else {
				writer.Write([]byte(fmt.Sprintf("%s %s %d %s\n", info.Mode().String(), info.ModTime().String(), info.Size(), entry.Name())))
			}
		}
		writer.Write([]byte(fmt.Sprintf("[%v] INFO 上传服务器端到存储 \n", time.Now().Format("2006-01-02 13:04:05"))))
		p, s, err := tools.UploadToOss(targetID)
		if err != nil {
			writer.Write([]byte(fmt.Sprintf("[%v] ERROR 上传到存储失败: %s\n", time.Now().Format("2006-01-02 13:04:05"), err.Error())))

		} else {
			record := upload.UploadRecords{
				GameServer: file,
				DestTarget: s,
				DestPath:   p,
				UploadTime: startTime,
			}
			config.DB.Create(&record)
		}
		writer.Write([]byte(fmt.Sprintf("[%v] Success 上传服务器端到存储成功  Path: %s \n", time.Now().Format("2006-01-02 13:04:05"), p)))
	}()
	return c.SendStream(reader)
}
func (u *UploadHandler) Handler_ShowUploadRecords(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize, _ := strconv.Atoi(c.Params("pageSize", "10"))
	var records []upload.UploadRecords
	var total int64
	if err := config.DB.Model(&records).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := config.DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    records,
	})
}
