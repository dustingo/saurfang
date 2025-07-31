package pkg

//
//import (
//	"bytes"
//	"context"
//	"encoding/json"
//	"fmt"
//	"log"
//	"net/http"
//	"os"
//	"saurfang/internal/config"
//	"saurfang/internal/models/dashboard"
//	"saurfang/internal/models/notice"
//	"saurfang/internal/models/task.go"
//	"saurfang/internal/service/taskservice"
//	"saurfang/internal/tools"
//
//	"strconv"
//	"time"
//
//	"github.com/hibiken/asynq"
//	"gorm.io/gorm"
//)
//
//type ScheduledJobProvider struct {
//	DB *gorm.DB
//}
//
//func NewScheduledJobProvider(db *gorm.DB) ScheduledJobProvider {
//	return ScheduledJobProvider{DB: db}
//}
//
//func (s ScheduledJobProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
//	ns := NewScheduledJobProvider(config.DB)
//	var jobs []task.CronJobs
//	if err := ns.DB.Where("task_status = ?", 0).Find(&jobs).Error; err != nil {
//		return nil, err
//	}
//	var configs []*asynq.PeriodicTaskConfig
//	for _, job := range jobs {
//		payload := []byte(strconv.Itoa(job.TaskID))
//		configs = append(configs, &asynq.PeriodicTaskConfig{Cronspec: job.Spec,
//			Task: asynq.NewTask(job.TaskType, payload, asynq.MaxRetry(-1)),
//			Opts: []asynq.Option{
//				asynq.Queue(config.SynqConfig.Queue),
//			},
//		})
//	}
//	return configs, nil
//}
//func TaskManagerSetup() {
//	go func() {
//		location, err := time.LoadLocation(config.SynqConfig.Location)
//		if err != nil {
//			log.Fatalln(err)
//			return
//		}
//		provider := NewScheduledJobProvider(config.DB)
//		mgr, err := asynq.NewPeriodicTaskManager(
//			asynq.PeriodicTaskManagerOpts{
//				RedisConnOpt:               asynq.RedisClientOpt{Addr: config.SynqConfig.Addr, Password: config.SynqConfig.Password, DB: config.SynqConfig.DB},
//				PeriodicTaskConfigProvider: provider,
//				SchedulerOpts:              &asynq.SchedulerOpts{Location: location},
//				SyncInterval:               time.Duration(config.SynqConfig.SyncInterval) * time.Second,
//			})
//		if err != nil {
//			log.Fatalln(err)
//			return
//		}
//		if err := mgr.Start(); err != nil {
//			log.Fatalln(err)
//			return
//		}
//	}()
//
//}
//
////// payload是运维任务的taskid
////func CronTaskHandler(ctx context.Context, at *asynq.Task) error {
////	execTime := time.Now()
////	payload := at.Payload()
////	taskID, err := strconv.Atoi(string(payload))
////	if err != nil {
////		return err
////	}
////	defer func(orm *gorm.DB, t time.Time, id int) {
////		if err := orm.Model(&task.CronJobs{}).Where("task_id = ?", taskID).Update("last_execution", t).Error; err != nil {
////			return
////		}
////		//orm.Exec("update saurfang_opstasks set last_execution =  ? where id = ? ;", t, taskID)
////	}(config.DB, execTime, taskID)
////	t := taskservice.NewOpsTaskService(config.DB)
////	taskInfo, err := t.ListByID(uint(taskID))
////	if err != nil {
////		return err
////	}
////	defer func(orm *gorm.DB, t string) {
////		var item dashboard.TaskDashboards
////		if err := orm.Where("task = ?", t).First(&item).Error; err != nil {
////			if err == gorm.ErrRecordNotFound {
////				orm.Create(&dashboard.TaskDashboards{Task: t, Count: 1})
////			}
////		} else {
////			item.Count += 1
////			orm.Model(&dashboard.TaskDashboards{}).Where("task = ?", t).Update("count", item.Count)
////		}
////	}(config.DB, taskInfo.Description)
////	hosts := taskInfo.Hosts
////	playbooks_keys := taskInfo.Playbooks
////	var cronJobs task.CronJobs
////	// 注意
////	// 由于删除是标记删除，因此在查询时必须加条件
////	if err := config.DB.Raw("select * from cron_jobs where task_id = ? and deleted_at is null;", taskID).Scan(&cronJobs).Error; err != nil {
////		return err
////	}
////	err = tools.RunAnsibleOpsPlaybooksByCron(hosts, playbooks_keys)
////	if err != nil {
////		config.DB.Model(&task.CronJobs{}).Where("task_id = ?", taskID).Update("task_status", 2)
////		if cronJobs.Ntfy {
////			go applyNotice(fmt.Sprintf("定时任务执行失败,任务ID: %s,任务名称: %s ", strconv.Itoa(taskID), cronJobs.TaskName), cronJobs.NtfyType, cronJobs.NtfyTarget, execTime)
////		}
////		return err
////	}
////	//0 未执行 1 执行成功 2 执行失败
////	config.DB.Model(&task.CronJobs{}).Where("task_id = ?", taskID).Update("task_status", 1)
////	if cronJobs.Ntfy {
////		go applyNotice(fmt.Sprintf("定时任务执行成功,任务ID: %s,任务名称: %s ", strconv.Itoa(taskID), cronJobs.TaskName), cronJobs.NtfyType, cronJobs.NtfyTarget, execTime)
////	}
////
////	return nil
////}
//
//func applyNotice(info string, ntype, nteam string, start time.Time) {
//	msg := notice.PushPayload{
//		Status:      "任务通知",
//		AlType:      ntype,
//		Team:        nteam,
//		Description: info,
//		Summary:     info,
//		StartsAt:    &start,
//	}
//	bf, err := json.Marshal(&msg)
//	if err != nil {
//
//		return
//	}
//	client := http.Client{}
//	req, err := http.NewRequest("POST", os.Getenv("NTFY_ADDR"), bytes.NewBuffer(bf))
//	if err != nil {
//
//		return
//	}
//	req.Header.Set("Content-Type", "application/json")
//	req.SetBasicAuth(os.Getenv("NTFY_USER"), os.Getenv("NTFY_PASSWORD"))
//	resp, err := client.Do(req)
//	if err != nil {
//		return
//	}
//	defer resp.Body.Close()
//	//body, _ := io.ReadAll(resp.Body)
//}
