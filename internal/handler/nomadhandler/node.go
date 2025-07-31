package nomadhandler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/models/nomadjob"
	"saurfang/internal/models/task.go"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/gofiber/fiber/v3"
	consulapi "github.com/hashicorp/consul/api"
	nomadapi "github.com/hashicorp/nomad/api"
)

type NomadHandler struct {
	base.NomadJobRepository
}

func NewNomadHandler(cli *consulapi.Client, ns string) *NomadHandler {
	client, err := config.NewNomadClient()
	if err != nil {
		slog.Error("NewNomadHandler err:", err)
		os.Exit(-1)
	}
	return &NomadHandler{
		base.NomadJobRepository{Consul: cli, Nomad: client, Ns: ns},
	}
}

// Handler_ListNomadNodes 列出所有node信息
func (n *NomadHandler) Handler_ListNomadNodes(ctx fiber.Ctx) error {
	cli, err := config.NewNomadClient()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": fmt.Sprintf("create nomad client error:%v", err),
		})
	}
	nodes, _, err := cli.Nodes().List(&nomadapi.QueryOptions{})
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": fmt.Sprintf("list nomad nodes error:%v", err),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    nodes,
	})
}

// Handler_ShowNomadJobs 展示Jobs的Group状态，变相等于进程状态
func (n *NomadHandler) Handler_ShowNomadJobs(ctx fiber.Ctx) error {
	jobType := ctx.Query("type", "")
	data, err := n.ShowNomadJobGroups(jobType)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": fmt.Sprintf("show job allocations error:%v", err),
		})
	}
	if data == nil {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  0,
			"message": "success",
			"data": fiber.Map{
				"items": []any{},
			},
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"items": data,
		},
	})
}

// Handler_ScaleJobs 伸缩group，此处使用神错达到启停的功能
func (n *NomadHandler) Handler_ScaleTaskGroup(ctx fiber.Ctx) error {
	var payload nomadjob.ScalePayload
	jobID := ctx.Params("job_id")
	ops := ctx.Query("ops")
	if err := ctx.Bind().Body(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": fmt.Sprintf("bind payload error:%v", err),
		})
	}
	evalID, err := n.ScaleTaskGroup(jobID, payload.Target, ops)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": fmt.Sprintf("scale job group error:%v", err),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    evalID,
	})
}

// Handler_ShowGroupSelect 执行单个group(进进程操作)时选择指定group
func (n *NomadHandler) Handler_ShowGroupSelect(ctx fiber.Ctx) error {
	jobID := ctx.Params("job_id")
	data, err := n.ShowGroupsForSelect(jobID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": fmt.Sprintf("show group select error:%v", err),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"options": data,
		},
	})
}

// Handler_DeployNomadOpsJob 执行nomad 运维任务(开关)
func (n *NomadHandler) Handler_DeployNomadOpsJob(ctx fiber.Ctx) error {
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")
	serverIDs := ctx.Query("server_ids")
	ops := ctx.Query("ops", "")
	keys := strings.Split(serverIDs, ",")
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		contents := make([]map[string]string, 0)
		kv := n.Consul.KV()
		// 获取配置内容
		for _, key := range keys {
			content := make(map[string]string)
			pair, _, err := kv.Get(tools.AddNamespace(key, n.Ns), &consulapi.QueryOptions{})
			if err != nil {
				writer.Write([]byte(fmt.Sprintf("data: [X] search config file failed. id: %s\n\n", key)))
				continue
			}
			if pair == nil {
				writer.Write([]byte(fmt.Sprintf("data: [X] config file not found. id: %s\n\n", key)))
				continue
			}
			content[key] = strings.ReplaceAll(string(pair.Value), "\r", "")
			contents = append(contents, content)
		}
		if len(contents) == 0 {
			writer.Write([]byte("data: [X] all config files were not found! check!check again!\n\n"))
			return
		}
		var wg sync.WaitGroup
		switch ops {
		case "stop":
			for _, content := range contents {
				for k, v := range content {
					job, err := n.Nomad.Jobs().ParseHCL(v, true)
					if err != nil {
						writer.Write([]byte(fmt.Sprintf("data: [X] parse job hcl config file failed. id: %s\n\n", k)))
						continue
					}
					wg.Add(1)
					go func(id string, serverID string) {
						defer wg.Done()
						res, _, err := n.Nomad.Jobs().Deregister(id, false, nil)
						if err != nil {
							writer.Write([]byte(fmt.Sprintf("data: [X] stop job failed. key: %s job: %s, error: %v, message: %s\n\n", serverID, id, err, res)))
							return
						}
						writer.Write([]byte(fmt.Sprintf("data: [√] stop job success. key: %s job: %s\n\n", serverID, id)))
						config.DB.Exec("UPDATE games set status = 0 where server_id = ?;", serverID)
					}(*job.ID, k)
				}
			}
			return
		case "start":
			for _, content := range contents {
				for k, v := range content {
					job, err := n.Nomad.Jobs().ParseHCL(v, true)
					if err != nil {
						writer.Write([]byte(fmt.Sprintf("data: [X] parse job hcl config file failed. id: %s\n\n", k)))
						continue
					}
					wg.Add(1)
					go func(j *nomadapi.Job, serverID string) {
						defer wg.Done()
						res, _, err := n.Nomad.Jobs().Register(j, nil)
						if err != nil {
							writer.Write([]byte(fmt.Sprintf("data: [X] deploy job failed. key: %s job: %s, error: %v\n\n", serverID, *j.ID, err)))
							return
						}
						writer.Write([]byte(fmt.Sprintf("data: [√] deploy job success. key: %s job: %s, evalID: %s\n\n", serverID, *j.ID, res.EvalID)))

						config.DB.Exec("UPDATE games set status = 1 where server_id = ?;", serverID)

					}(job, k)
				}
			}
			return
		default:
			writer.Write([]byte(fmt.Sprintf("data: [X] unknown operation type: %s", ops)))
		}
		wg.Wait()
		writer.Write([]byte("data: [√] All operations completed\n\n"))
	}()

	return ctx.SendStream(reader)
}

// Handler_DeployNomadJob 执行nomad一次性任务dispatch
func (n *NomadHandler) Handler_DeployNomadJob(ctx fiber.Ctx) error {
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")
	serverIDs := ctx.Query("server_ids")
	keys := strings.Split(serverIDs, ",")
	reader, writer := io.Pipe()
	results := make(chan task.JobResult, len(keys))
	go func() {
		defer writer.Close()
		contents := make([]map[string]string, 0)
		kv := n.Consul.KV()
		// 获取配置内容
		for _, key := range keys {
			content := make(map[string]string)
			pair, _, err := kv.Get(tools.AddNamespace(key, n.Ns), nil)
			if err != nil {
				writer.Write([]byte(fmt.Sprintf("data: [X] search config file failed. id: %s\n\n", key)))
				continue
			}
			if pair == nil {
				writer.Write([]byte(fmt.Sprintf("data: [X] config file not found. id: %s\n\n", key)))
				continue
			}
			content[key] = strings.TrimRight(string(pair.Value), "\r")
			contents = append(contents, content)
		}
		if len(contents) == 0 {
			writer.Write([]byte(fmt.Sprintf("data: [X] all config files were not found! check!check again!\n\n")))
			return
		}
		var wg sync.WaitGroup
		for _, content := range contents {
			for k, v := range content {
				job, err := n.Nomad.Jobs().ParseHCL(v, true)
				if err != nil {
					writer.Write([]byte(fmt.Sprintf("data: [X] parse job hcl config file failed. id: %s\n\n", k)))
					continue
				}
				wg.Add(1)
				go func(j *nomadapi.Job, serverID string) {
					defer wg.Done()
					_, _, err := n.Nomad.Jobs().Register(j, nil)
					if err != nil {
						writer.Write([]byte(fmt.Sprintf("data: [X] register dispatch job failed. key: %s job: %s, error: %v\n\n", serverID, *j.ID, err)))
						return
					}
					meta := make(map[string]string)
					meta["EXEC_TIME"] = time.Now().String()
					res, _, err := n.Nomad.Jobs().Dispatch(*j.ID, meta, []byte(time.Now().String()), "", nil)
					//res, _, err := n.Nomad.Jobs().Register(j, nil)
					if err != nil {
						fmt.Println("dispatch error", err.Error())
						writer.Write([]byte(fmt.Sprintf("data: [X] deploy job failed. key: %s job: %s, error: %v\n\n", serverID, *j.ID, err)))
						waitEvalCompletion(n.Nomad, serverID, res.EvalID, 120*time.Second, writer, results)
						return
					}
					waitEvalCompletion(n.Nomad, serverID, res.EvalID, 120*time.Second, writer, results)
					writer.Write([]byte(fmt.Sprintf("data: [√] deploy job success. key: %s job: %s, evalID: %s\n\n", serverID, *j.ID, res.EvalID)))
				}(job, k)
			}
		}
		// 等待所有异步操作完成
		wg.Wait()
		close(results)
		writer.Write([]byte("data: [√] All operations completed\n\n"))
		writer.Write([]byte(fmt.Sprintln(tools.FormatLine("Job [result]", "*"))))
		headers := []string{"EvalID", "JobID", "Type", "Status", "TriggeredBy"}
		drawTable(results, headers, writer)
	}()
	return ctx.SendStream(reader)
}

// Handler_PurgeNomadJob 清除nomad job
func (n *NomadHandler) Handler_PurgeNomadJob(ctx fiber.Ctx) error {
	ids := ctx.Query("job_ids")
	var wg sync.WaitGroup
	for _, id := range strings.Split(ids, ",") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, _, err := n.Nomad.Jobs().Deregister(id, true, nil)
			if err != nil {
				slog.Error("purge job failed", "id", id, "err", err, "message", res)
				return
			}
		}()
	}
	wg.Wait()
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// waitEvalCompletion 轮询eval状态
func waitEvalCompletion(client *nomadapi.Client, serverID, evalID string, timeout time.Duration, writer *io.PipeWriter, ch chan task.JobResult) {
	defer func() {
		if err := recover(); err != nil {
			slog.Warn("eval completion panic", "err", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			writer.Write([]byte(fmt.Sprintf("timeout waiting for eval %s to complete,serverID: %s\n", evalID, serverID)))
		case <-ticker.C:
			eval, _, err := client.Evaluations().Info(evalID, nil)
			if err != nil {
				writer.Write([]byte(fmt.Sprintf("failed to get job eval info. serverID: %s, error: %v\n\n", serverID, err)))
				return
			}
			switch eval.Status {
			case "complete", "failed", "cancelled", "blocked":
				ch <- task.JobResult{
					EvalID:      eval.ID,
					JobID:       eval.JobID,
					Type:        eval.Type,
					TriggeredBy: eval.TriggeredBy,
					Status:      eval.Status,
				}
				return
			case "pending":
				// 继续等待
			default:
				// 其他状态根据需求处理，或继续等待
			}
		}
	}
}

// 格式化输出结果
func drawTable(ch chan task.JobResult, headers []string, writer *io.PipeWriter) {
	var data [][]string
	for result := range ch {
		item := []string{result.EvalID, result.JobID, result.Type, result.Status, result.TriggeredBy}
		data = append(data, item)
	}
	writeTable(writer, headers, data)
}
func writeTable(w io.Writer, headers []string, data [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	// 表头
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	// 数据行
	for _, row := range data {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	tw.Flush()
}
