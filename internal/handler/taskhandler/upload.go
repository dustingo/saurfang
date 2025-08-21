package taskhandler

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"saurfang/internal/config"
	"saurfang/internal/models/notify"
	"saurfang/internal/models/upload"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools"
	"saurfang/internal/tools/ntfy"
	"saurfang/internal/tools/pkg"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

type UploadHandler struct {
	base.BaseGormRepository[upload.UploadRecord]
}

// Handler_ShowServerPackage 显示服务器端文件列表
func (u *UploadHandler) Handler_ShowServerPackage(c fiber.Ctx) error {
	var files []upload.FileInfo
	entries, err := os.ReadDir(os.Getenv("SERVER_PACKAGE_SRC_PATH"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "read dir error", err.Error(), fiber.Map{})
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
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", files)
}

// Handler_UploadServerPackage 上传服务器端文件
func (u *UploadHandler) Handler_UploadServerPackage(c fiber.Ctx) error {
	u.setSSEHeaders(c)
	file := c.Query("file")
	targetID, _ := strconv.Atoi(c.Query("target"))
	startTime := time.Now()
	var successCount, failCount int
	var successJobs, failedJobs []string
	var mu sync.Mutex
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		if _, err := os.Stat(path.Join(os.Getenv("SERVER_PACKAGE_SRC_PATH"), file)); err != nil {
			writer.Write([]byte(fmt.Sprintf("[%v] ERROR 缺少服务器端文件\n", time.Now().Format("2006-01-02 13:04:05"))))
			u.recordFailedJob(&mu, &failCount, &failedJobs, file)
			return
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
				u.recordFailedJob(&mu, &failCount, &failedJobs, file)
				return
			}
		}
		writer.Write([]byte(fmt.Sprintf("[%v] Success 清空目录成功\n", time.Now().Format("2006-01-02 13:04:05"))))
		writer.Write([]byte(fmt.Sprintf("[%v] INFO 正在解压服务器端 %s 到 %s\n", time.Now().Format("2006-01-02 13:04:05"), file, os.Getenv("SERVER_PACKAGE_DEST_PATH"))))
		if err := tools.SafeUnzip(path.Join(os.Getenv("SERVER_PACKAGE_SRC_PATH"), file), os.Getenv("SERVER_PACKAGE_DEST_PATH")); err != nil {
			writer.Write([]byte(fmt.Sprintf("[%v] ERROR 解压服务器端失败 %s\n", time.Now().Format("2006-01-02 13:04:05"), err.Error())))
			u.recordFailedJob(&mu, &failCount, &failedJobs, file)
			return
		}
		entries, err := os.ReadDir(os.Getenv("SERVER_PACKAGE_DEST_PATH"))
		if err != nil {
			writer.Write([]byte(fmt.Sprintf("[%v] ERROR 获取资源列表失败 %s\n", time.Now().Format("2006-01-02 13:04:05"), err.Error())))
			u.recordFailedJob(&mu, &failCount, &failedJobs, file)
			return
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
			u.recordFailedJob(&mu, &failCount, &failedJobs, file)
			return
		} else {
			record := upload.UploadRecord{
				GameServer: file,
				DestTarget: s,
				DestPath:   p,
				UploadTime: startTime,
			}
			config.DB.Create(&record)
		}
		writer.Write([]byte(fmt.Sprintf("[%v] Success 上传服务器端到存储成功  Path: %s \n", time.Now().Format("2006-01-02 13:04:05"), p)))
		u.recordSuccessJob(&mu, &successCount, &successJobs, file)
		ntfy.PublishNotification(notify.EventTypeUpload, fmt.Sprintf("upload %s", file), successJobs, failedJobs, successCount, failCount)
	}()
	return c.SendStream(reader)
}

// Handler_ShowUploadRecords 显示上传记录
func (u *UploadHandler) Handler_ShowUploadRecords(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize, _ := strconv.Atoi(c.Params("pageSize", "10"))
	var records []upload.UploadRecord
	var total int64
	if err := config.DB.Model(&records).Count(&total).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "fail to count records", err.Error(), fiber.Map{})
	}
	if err := config.DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "fail to find records", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", records)
}

// setSSEHeaders 设置SSE响应头
func (*UploadHandler) setSSEHeaders(ctx fiber.Ctx) {
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")
	ctx.Set("Access-Control-Allow-Origin", "*")
	ctx.Set("Access-Control-Allow-Headers", "Cache-Control")
}

// recordFailedJob 记录失败的任务
func (n *UploadHandler) recordFailedJob(mu *sync.Mutex, failCount *int, failedJobs *[]string, key string) {
	mu.Lock()
	defer mu.Unlock()
	*failCount++
	*failedJobs = append(*failedJobs, key)
}

// recordSuccessJob 记录成功的任务
func (n *UploadHandler) recordSuccessJob(mu *sync.Mutex, successCount *int, successJobs *[]string, key string) {
	mu.Lock()
	defer mu.Unlock()
	*successCount++
	*successJobs = append(*successJobs, key)
}
