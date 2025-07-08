package taskhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/models/dashboard"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/models/task.go"
	"saurfang/internal/service/datasrcservice"
	"saurfang/internal/service/taskservice"
	"saurfang/internal/tools"
	"saurfang/internal/tools/pkg"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DeployHandler struct {
	taskservice.DeployService
}

func NewDeployHandler(svc *taskservice.DeployService) *DeployHandler {
	return &DeployHandler{*svc}
}

/*
发布分为程序发布和配置发布
*/
// Handler_CreateDeployTask 创建程序发布任务
func (d *DeployHandler) Handler_CreateDeployTask(c fiber.Ctx) error {
	var payload task.PublishTaskParams
	var task task.SaurfangPublishtasks
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	ds := datasrcservice.NewDataSourceService(config.DB)
	data, err := ds.Service_ShowDataSourceByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := c.Bind().Body(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	task.ID = uint(id)
	task.SourceLabel = data.Label
	task.Become = int(payload.Become)
	task.BecomeUser = payload.BecomeUser
	task.Comment = payload.Comment
	if err := d.Create(&task); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})

	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  1,
		"message": "success",
	})
}
func (d *DeployHandler) Handler_DeleteDeployTask(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := d.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}
func (d *DeployHandler) Handler_ShowDeployTask(c fiber.Ctx) error {
	tasks, err := d.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    tasks,
	})
}
func (d *DeployHandler) Handler_ShowDeployTaskByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	task, err := d.ListByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    task,
	})
}
func (d *DeployHandler) Handler_ShowDeployPerPage(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Params("page", "1"))
	pageSize, _ := strconv.Atoi(c.Params("pageSize", "10"))
	tasks, total, err := d.ListPerPage(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    tasks,
		"total":   total,
	})

}

// Handler_RunGameDeployTask 执行程序发布任务
func (d *DeployHandler) Handler_RunGameDeployTask(c fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	hosts := c.Query("hosts")
	taskID, _ := strconv.Atoi(c.Query("tid"))
	taskSvc := taskservice.NewDeployService(config.DB)
	task, err := taskSvc.ListByID(uint(taskID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	dssvc := datasrcservice.NewDataSourceService(config.DB)
	datasource, err := dssvc.ListByID(uint(task.SourceID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	realhosts, err := tools.SortHosts(hosts)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	//realHosts: map[server_id][]string{ip}
	var inventoryHosts []string
	for _, hosts := range realhosts {
		for _, host := range hosts {
			inventoryHosts = append(inventoryHosts, host)
		}
	}
	// 游戏服配置
	configs, err := config.Etcd.Get(context.Background(), os.Getenv("GAME_CONFIG_NAMESPACE"), clientv3.WithPrefix())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	var keys []string //存储在etcd中所有的key
	for _, kv := range configs.Kvs {
		keys = append(keys, string(kv.Key))
	}
	// keys: [game_config/server_id]
	sort.Strings(keys) //对keys进行排序
	var lastConfigsMap map[string]serverconfig.Configs = make(map[string]serverconfig.Configs)
	var gameconfigs serverconfig.GameConfigs
	for serverId, ips := range realhosts {
		if tools.ContainsKeysSorted(keys, fmt.Sprintf("%s/%s", os.Getenv("GAME_CONFIG_NAMESPACE"), serverId)) {
			for _, ip := range ips {
				cnf, err := config.Etcd.Get(context.Background(), fmt.Sprintf("%s/%s", os.Getenv("GAME_CONFIG_NAMESPACE"), serverId))
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"status":  1,
						"message": err.Error(),
					})
				}
				// 实际正常只有一个值kv

				for _, kv := range cnf.Kvs {
					if err := json.Unmarshal(kv.Value, &gameconfigs); err != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"status":  1,
							"message": err.Error(),
						})
					}
					for _, cnf := range gameconfigs.Configs {
						if cnf.IP == ip {
							// 将所有逻辑服中服务器与配置里正常配置的服务器配置重新存储到map
							lastConfigsMap[fmt.Sprintf("%s-%s", serverId, cnf.SvcName)] = cnf
						}
					}
				}
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  1,
				"message": "<UNK>",
			})
		}
	}
	reader, writer := io.Pipe()
	go tools.RunAnsibleDeployPlaybooks(strings.Join(inventoryHosts, ","), datasource, task, lastConfigsMap, *writer)
	return c.SendStream(reader)
}

// Handler_RunOpsTask 执行常规ansible playbook任务
func (d *DeployHandler) Handler_RunOpsTask(c fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	execTime := time.Now()
	taskID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	defer func(orm *gorm.DB, t time.Time, id int) {
		if err := orm.Model(&task.SaurfangOpstasks{}).Where("id = ?", taskID).Update("last_execution", t).Error; err != nil {
			log.Println("update ops task last_execution failed")
		}
		//orm.Exec("update saurfang_opstasks set last_execution =  ? where id = ? ;", t, taskID)
	}(config.DB, execTime, taskID)
	taskSvc := taskservice.NewOpsTaskService(config.DB)
	taskInfo, err := taskSvc.ListByID(uint(taskID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	defer func(orm *gorm.DB, t string) {
		var item dashboard.SaurfangTaskdashboards
		if err := orm.Where("task = ?", t).First(&item).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				orm.Create(&dashboard.SaurfangTaskdashboards{Task: t, Count: 1})
			}
		} else {
			item.Count += 1
			orm.Model(&dashboard.SaurfangTaskdashboards{}).Where("task = ?", t).Update("count", item.Count)
		}
	}(config.DB, taskInfo.Description)
	hosts := taskInfo.Hosts
	playbooks_keys := taskInfo.Playbooks
	reader, writer := io.Pipe()
	go tools.RunAnsibleOpsPlaybooks(hosts, playbooks_keys, *writer)
	return c.SendStream(reader)
}

// Handler_RunConfigDeployTask 配置发布
func (d *DeployHandler) Handler_RunConfigDeployTask(c fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	hosts := c.Query("hosts")
	reader, writer := io.Pipe()
	go func(target string) {
		defer writer.Close()
		realHosts, err := tools.SortHosts(target)
		if err != nil {
			writer.Write([]byte("run ops task failed: " + err.Error()))
			return
		}
		configs, err := config.Etcd.Get(context.Background(), os.Getenv("GAME_CONFIG_NAMESPACE"), clientv3.WithPrefix())
		if err != nil {
			writer.Write([]byte("run ops task failed: " + err.Error()))
			return
		}
		var keys []string //存储在etcd中所有的key
		for _, kv := range configs.Kvs {
			keys = append(keys, string(kv.Key))
		}
		sort.Strings(keys) //对keys进行排序
		var lastConfigsMap map[string]serverconfig.Configs = make(map[string]serverconfig.Configs)
		writer.Write([]byte(tools.FormatLine("TASK [服务器ServerID]", "*")))
		for serverId, ips := range realHosts {
			if tools.ContainsKeysSorted(keys, fmt.Sprintf("%s/%s", os.Getenv("GAME_CONFIG_NAMESPACE"), serverId)) {
				writer.Write([]byte(serverId + "\n"))
				for _, ip := range ips {
					res, err := config.Etcd.Get(context.Background(), fmt.Sprintf("%s/%s", os.Getenv("GAME_CONFIG_NAMESPACE"), serverId))
					if err != nil {
						writer.Write([]byte("run ops task failed: " + err.Error()))
					}
					// 实际正常只有一个值kv
					for _, kv := range res.Kvs {
						var gameconfigs serverconfig.GameConfigs
						if err := json.Unmarshal(kv.Value, &gameconfigs); err != nil {
							writer.Write([]byte("run ops task failed: " + err.Error()))
						}
						for _, cnf := range gameconfigs.Configs {
							if cnf.IP == ip {
								// 将所有逻辑服中服务器与配置里正常配置的服务器配置重新存储到map
								lastConfigsMap[fmt.Sprintf("%s-%s", serverId, cnf.SvcName)] = cnf
							}
						}
					}
				}
			} else {
				writer.Write([]byte("record not found," + serverId + "\n"))
			}
		}
		writer.Write([]byte(tools.FormatLine("TASK [服务器IP]", "*")))
		for ip, cnf := range lastConfigsMap {
			writer.Write([]byte(fmt.Sprintf("%s\t %s   %v\t \n", ip, cnf.ConfigFile, cnf.Vars)))
		}
		writer.Write([]byte(tools.FormatLine("TASK [同步配置]", "*")))
		concurrency, _ := strconv.Atoi(os.Getenv("EXEC_CONCURRENCY"))
		results := make(chan task.TaskStatus, concurrency)
		var wg sync.WaitGroup
		sem := make(chan struct{}, concurrency)
		for _, cnf := range lastConfigsMap {
			wg.Add(1)
			sem <- struct{}{}
			go func(config serverconfig.Configs) {
				defer wg.Done()
				defer func() { <-sem }()
				taskResult := task.TaskStatus{}
				err := pkg.ProcessServer(config)
				if err != nil {
					writer.Write([]byte(fmt.Sprintf("[failed] %s\t  %s\n", config.IP, err.Error())))
					taskResult.Status = "failure"
					taskResult.ErrorMessage = err.Error()
					results <- taskResult
					return
				} else {
					taskResult.Status = "success"
				}
				writer.Write([]byte(fmt.Sprintf("[ok] %s\n", config.IP)))
				results <- taskResult
			}(cnf)
		}
		wg.Wait()
		close(results)
		writer.Write([]byte(tools.FormatLine("TASK [result]", "*")))
		writer.Write([]byte(tools.TaskReport(results)))
	}(hosts)
	return c.SendStream(reader)
}
