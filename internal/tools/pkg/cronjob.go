package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"saurfang/internal/config"
	"saurfang/internal/models/notify"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/models/task"
	"saurfang/internal/tools"
	"saurfang/internal/tools/ntfy"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type ScheduledJobProvider struct {
	DB *gorm.DB
}

func NewScheduledJobProvider(db *gorm.DB) ScheduledJobProvider {
	return ScheduledJobProvider{DB: db}
}

func (s ScheduledJobProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	// 获取扩展的 CronJobs 配置（CustomTask 和 Server 操作）
	extendedConfigs, err := s.getExtendedCronJobConfigs()
	if err != nil {
		return nil, err
	}

	return extendedConfigs, nil
}

// getExtendedCronJobConfigs 获取扩展的 CronJobs 配置
func (s ScheduledJobProvider) getExtendedCronJobConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var jobs []task.CronJobs
	if err := s.DB.Where("task_status = ? AND task_type IN (?, ?)", 0, task.TaskTypeCustom, task.TaskTypeServer).Find(&jobs).Error; err != nil {
		return nil, err
	}

	var configs []*asynq.PeriodicTaskConfig
	for _, job := range jobs {
		// 创建包含任务类型和ID的payload
		payload := map[string]interface{}{
			"cron_job_id": job.ID,
			"task_type":   job.TaskType,
		}

		// 根据任务类型添加特定信息
		switch job.TaskType {
		case task.TaskTypeCustom:
			if job.CustomTaskID != nil {
				payload["custom_task_id"] = *job.CustomTaskID
			}
		case task.TaskTypeServer:
			if job.ServerIDs != "" {
				payload["server_ids"] = strings.Split(job.ServerIDs, ",")
				payload["server_operation"] = job.ServerOperation
			}
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal payload for job %d: %v", job.ID, err)
			continue
		}

		configs = append(configs, &asynq.PeriodicTaskConfig{
			Cronspec: job.Spec,
			Task:     asynq.NewTask(job.TaskType, payloadBytes, asynq.MaxRetry(-1)),
			Opts: []asynq.Option{
				asynq.Queue(config.SynqConfig.Queue),
			},
		})
	}
	return configs, nil
}

func TaskManagerSetup() {
	go func() {
		location, err := time.LoadLocation(config.SynqConfig.Location)
		if err != nil {
			log.Fatalln(err)
			return
		}
		provider := NewScheduledJobProvider(config.DB)
		mgr, err := asynq.NewPeriodicTaskManager(
			asynq.PeriodicTaskManagerOpts{
				RedisConnOpt:               asynq.RedisClientOpt{Addr: config.SynqConfig.Addr, Password: config.SynqConfig.Password, DB: config.SynqConfig.DB},
				PeriodicTaskConfigProvider: provider,
				SchedulerOpts:              &asynq.SchedulerOpts{Location: location},
				SyncInterval:               time.Duration(config.SynqConfig.SyncInterval) * time.Second,
			})
		if err != nil {
			log.Fatalln(err)
			return
		}
		if err := mgr.Start(); err != nil {
			log.Fatalln(err)
			return
		}
	}()
}

// ServerOperationHandler 处理游戏服务器操作
func ServerOperationHandler(ctx context.Context, at *asynq.Task) error {
	execTime := time.Now()
	var payload map[string]interface{}
	if err := json.Unmarshal(at.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	cronJobID, ok := payload["cron_job_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid cron_job_id in payload")
	}

	serverIDsInterface, ok := payload["server_ids"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid server_ids in payload")
	}

	serverOperation, ok := payload["server_operation"].(string)
	if !ok {
		return fmt.Errorf("invalid server_operation in payload")
	}

	// 转换 serverIDs
	var serverIDs []string
	for _, id := range serverIDsInterface {
		if strID, ok := id.(string); ok {
			serverIDs = append(serverIDs, strID)
		}
	}
	// 更新最后执行时间
	defer func() {
		if err := config.DB.Model(&task.CronJobs{}).Where("id = ?", uint(cronJobID)).Update("last_execution", execTime).Error; err != nil {
			log.Printf("Failed to update last_execution: %v", err)
		}
	}()

	// 获取 CronJob 信息用于通知
	var cronJob task.CronJobs
	if err := config.DB.First(&cronJob, uint(cronJobID)).Error; err != nil {
		log.Printf("Failed to get cron job info: %v", err)
	}
	// 创建 nomad 客户端
	// nomad, err := config.NewNomadClient()
	// if err != nil {
	// 	return fmt.Errorf("failed to create nomad client: %v", err)
	// }
	//消息
	var successCount, failCount int
	var successJobs, failedJobs []string
	// 执行服务器操作
	// successCount := 0
	// failedCount := 0
	var errors []string
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, serverID := range serverIDs {
		wg.Add(1)
		go func(serverID string, failedCount *int, successCount *int, errors *[]string, wg *sync.WaitGroup, mu *sync.Mutex, successJobs, failedJobs *[]string) {
			//go func(serverID string, failedCount *int, successCount *int, errors *[]string, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wg.Done()
			// 从 Consul 获取服务器配置
			configKey := tools.AddNamespace(serverID, os.Getenv("GAME_NOMAD_JOB_NAMESPACE"))
			gameConfig, _, err := config.ConsulCli.KV().Get(configKey, nil)
			if err != nil {
				mu.Lock()
				*errors = append(*errors, fmt.Sprintf("Failed to get config for server %s: %v", serverID, err))
				*failedCount++
				*failedJobs = append(*failedJobs, serverID)
				mu.Unlock()
				return
			}
			if gameConfig == nil || gameConfig.Value == nil {
				mu.Lock()
				*errors = append(*errors, fmt.Sprintf("No config found for server %s", serverID))
				*failedCount++
				*failedJobs = append(*failedJobs, serverID)
				mu.Unlock()
				return
			}
			// 将获取到的kv转换成结构体
			gameConfigData := &serverconfig.GameConfig{
				Key:     gameConfig.Key,
				Setting: strings.ReplaceAll(string(gameConfig.Value), "\r", ""),
			}
			// 根据操作类型执行相应命令
			switch serverOperation {
			case task.ServerOpStart:
				defer func() {
					if r := recover(); r != nil {
						slog.Error("panic in deploy nomad ops job", "panic", r)
					}
				}()
				job, err := config.NomadCli.Jobs().ParseHCL(gameConfigData.Setting, true)
				if err != nil {
					mu.Lock()
					*errors = append(*errors, fmt.Sprintf("Failed to parse job hcl config file: %v", err))
					*failedCount++
					*failedJobs = append(*failedJobs, serverID)
					mu.Unlock()
					return
				}
				_, _, err = config.NomadCli.Jobs().Register(job, nil)
				if err != nil {
					mu.Lock()
					*errors = append(*errors, fmt.Sprintf("Failed to register job: %v", err))
					*failedCount++
					*failedJobs = append(*failedJobs, serverID)
					mu.Unlock()
					return
				}
				mu.Lock()
				*successCount++
				*successJobs = append(*successJobs, serverID)
				mu.Unlock()
				slog.Info("start nomad ops job success", "server_id", serverID)
				config.DB.Exec("UPDATE games set status = 1 where server_id = ?;", serverID)
				return
			case task.ServerOpStop:
				defer func() {
					if r := recover(); r != nil {
						slog.Error("panic in deploy nomad ops job", "panic", r)
					}
				}()
				job, err := config.NomadCli.Jobs().ParseHCL(gameConfigData.Setting, true)
				if err != nil {
					mu.Lock()
					*errors = append(*errors, fmt.Sprintf("Failed to parse job hcl config file: %v", err))
					*failedCount++
					*failedJobs = append(*failedJobs, serverID)
					mu.Unlock()
					return
				}
				_, _, err = config.NomadCli.Jobs().Deregister(*job.ID, false, nil)
				if err != nil {
					mu.Lock()
					*errors = append(*errors, fmt.Sprintf("Failed to register job: %v", err))
					*failedCount++
					*failedJobs = append(*failedJobs, serverID)
					mu.Unlock()
					return
				}
				mu.Lock()
				*successCount++
				*successJobs = append(*successJobs, serverID)
				mu.Unlock()
				slog.Info("stop nomad ops job success", "server_id", serverID)
				config.DB.Exec("UPDATE games set status = 0 where server_id = ?;", serverID)
				return
			}
		}(serverID, &failCount, &successCount, &errors, &wg, &mu, &successJobs, &failedJobs)
	}
	wg.Wait()
	// 如果有错误，返回所有错误信息
	ntfy.PublishNotification(notify.EventTypeCronJob, fmt.Sprintf("game %s", cronJob.TaskName), successJobs, failedJobs, successCount, failCount)
	if len(errors) > 0 {
		return fmt.Errorf("server operation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
