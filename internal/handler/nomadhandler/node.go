package nomadhandler

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"saurfang/internal/config"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/nomadjob"
	"saurfang/internal/models/notify"
	"saurfang/internal/models/task"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools"
	"saurfang/internal/tools/ntfy"
	"saurfang/internal/tools/pkg"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/gofiber/fiber/v3"
	consulapi "github.com/hashicorp/consul/api"
	nomadapi "github.com/hashicorp/nomad/api"
)

// 错误常量
const (
	statusOffline = 0
	statusOnline  = 1
)

type NomadHandler struct {
	base.NomadJobRepository
}

func NewNomadHandler(cli *consulapi.Client, ns string) *NomadHandler {
	client, err := config.NewNomadClient()
	if err != nil {
		slog.Error("create nomad client error", "error", err)
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
		return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "create nomad client error", err.Error(), nil)
	}
	nodes, _, err := cli.Nodes().List(&nomadapi.QueryOptions{})
	if err != nil {
		return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "list nomad nodes error", err.Error(), nil)
	}
	return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", nodes)
}

// Handler_ListNomadNodesForSelect 选择nomad node key:name(status) value:name
func (n *NomadHandler) Handler_ListNomadNodesForSelect(ctx fiber.Ctx) error {
	nodes, _, err := n.Nomad.Nodes().List(&nomadapi.QueryOptions{})
	if err != nil {
		return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "list nomad nodes error", err.Error(), nil)
	}
	var ops []amis.AmisOptionsGeneric[string]
	var op amis.AmisOptionsGeneric[string]
	for _, node := range nodes {
		op.Label = fmt.Sprintf("%s(%s)", node.Name, node.Status)
		op.Value = node.Name
		ops = append(ops, op)
	}
	return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", fiber.Map{
		"options": ops,
	})
}

// Handler_ShowNomadJobs 展示Jobs的Group状态，变相等于进程状态
func (n *NomadHandler) Handler_ShowNomadJobs(ctx fiber.Ctx) error {

	jobType := ctx.Query("type", "")
	data, err := n.ShowNomadJobGroups(jobType)
	if err != nil {
		return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "show job allocations error", err.Error(), nil)
	}
	if data == nil {
		return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", fiber.Map{
			"items": []any{},
		})
	}
	return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", fiber.Map{
		"items": data,
	})
}

// Handler_ScaleJobs 伸缩group，此处使用神错达到启停的功能
func (n *NomadHandler) Handler_ScaleTaskGroup(ctx fiber.Ctx) error {
	var payload nomadjob.ScalePayload
	jobID := ctx.Params("job_id")
	ops := ctx.Query("ops")
	if err := ctx.Bind().Body(&payload); err != nil {
		return pkg.NewAppResponse(ctx, fiber.StatusBadRequest, 1, "bind payload error", err.Error(), nil)
	}
	evalID, err := n.ScaleTaskGroup(jobID, payload.Target, ops)
	if err != nil {
		return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "scale job group error", err.Error(), nil)
	}
	return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", evalID)
}

// Handler_ShowGroupSelect 执行单个group(进进程操作)时选择指定group
func (n *NomadHandler) Handler_ShowGroupSelect(ctx fiber.Ctx) error {
	jobID := ctx.Params("job_id")
	if jobID == "" {
		return pkg.NewAppResponse(ctx, fiber.StatusBadRequest, 1, "job_id is required", "", nil)
	}
	data, err := n.ShowGroupsForSelect(jobID)
	if err != nil {
		return pkg.NewAppResponse(ctx, fiber.StatusInternalServerError, 1, "show group select error", err.Error(), nil)
	}
	return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", fiber.Map{
		"options": data,
	})
}

// Handler_DeployNomadOpsJob 执行nomad 运维任务(开关)
func (n *NomadHandler) Handler_DeployNomadOpsJob(ctx fiber.Ctx) error {
	n.setSSEHeaders(ctx)
	serverIDs := ctx.Query("server_ids")
	ops := ctx.Query("ops", "")
	var successCount, failCount int
	var successJobs, failedJobs []string
	// 消息通知
	if serverIDs == "" {
		return pkg.NewAppResponse(ctx, fiber.StatusBadRequest, 1, "server_ids is required", "", nil)
	}
	if ops == "" {
		return pkg.NewAppResponse(ctx, fiber.StatusBadRequest, 1, "ops is required", "", nil)
	}
	keys := strings.Split(serverIDs, ",")
	messageChan := make(chan string, 100)
	var mu sync.Mutex
	go func() {
		defer close(messageChan)
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in deploy nomad ops job", "panic", r)
				messageChan <- fmt.Sprintf("data: [X] Internal error occurred: %v\n\n", r)
			}
		}()
		contents := make([]map[string]string, 0)
		kv := n.Consul.KV()
		for _, key := range keys {
			content := make(map[string]string)
			pair, _, err := kv.Get(tools.AddNamespace(key, n.Ns), &consulapi.QueryOptions{})
			if err != nil {
				messageChan <- fmt.Sprintf("data: [X] search config file failed. id: %s\n\n", key)
				n.recordFailedJob(&mu, &failCount, &failedJobs, key)
				continue
			}
			if pair == nil {
				messageChan <- fmt.Sprintf("data: [X] config file not found. id: %s\n\n", key)
				n.recordFailedJob(&mu, &failCount, &failedJobs, key)
				continue
			}
			content[key] = strings.ReplaceAll(string(pair.Value), "\r", "")
			contents = append(contents, content)
		}
		if len(contents) == 0 {
			messageChan <- "data: [X] all config files were not found! check!check again!\n\n"
			return
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		switch ops {
		case "stop":
			for _, content := range contents {
				for k, v := range content {
					job, err := n.Nomad.Jobs().ParseHCL(v, true)
					if err != nil {
						messageChan <- fmt.Sprintf("data: [X] parse job hcl config file failed. id: %s\n\n", k)
						n.recordFailedJob(&mu, &failCount, &failedJobs, k)
						continue
					}
					wg.Add(1)
					go func(id string, serverID string) {
						defer wg.Done()
						res, _, err := n.Nomad.Jobs().Deregister(id, false, nil)
						if err != nil {
							messageChan <- fmt.Sprintf("data: [X] stop job failed. key: %s job: %s, error: %v, message: %s\n\n", serverID, id, err, res)
							n.recordFailedJob(&mu, &failCount, &failedJobs, k)
							return
						}
						messageChan <- fmt.Sprintf("data: [√] stop job success. key: %s job: %s\n\n", serverID, id)
						mu.Lock()
						n.updateGameStatus(serverID, statusOffline)
						mu.Unlock()
						n.recordSuccessJob(&mu, &successCount, &successJobs, k)
					}(*job.ID, k)
				}
			}
		case "start":
			for _, content := range contents {
				for k, v := range content {
					job, err := n.Nomad.Jobs().ParseHCL(v, true)
					if err != nil {
						messageChan <- fmt.Sprintf("data: [X] parse job hcl config file failed. id: %s\n\n", k)
						n.recordFailedJob(&mu, &failCount, &failedJobs, k)
						continue
					}
					wg.Add(1)
					go func(j *nomadapi.Job, serverID string) {
						defer wg.Done()
						res, _, err := n.Nomad.Jobs().Register(j, nil)
						if err != nil {
							messageChan <- fmt.Sprintf("data: [X] deploy job failed. key: %s job: %s, error: %v\n\n", serverID, *j.ID, err)
							n.recordFailedJob(&mu, &failCount, &failedJobs, k)
							return
						}
						messageChan <- fmt.Sprintf("data: [√] deploy job success. key: %s job: %s, evalID: %s\n\n", serverID, *j.ID, res.EvalID)
						mu.Lock()
						n.updateGameStatus(serverID, statusOnline)
						mu.Unlock()
						n.recordSuccessJob(&mu, &successCount, &successJobs, k)
					}(job, k)
				}
			}
		default:
			messageChan <- fmt.Sprintf("data: [X] unknown operation type: %s\n\n", ops)
		}
		wg.Wait()
		ntfy.PublishNotification(notify.EventTypeGameOps, fmt.Sprintf("game %s", ops), successJobs, failedJobs, successCount, failCount)
		messageChan <- "data: [√] All operations completed\n\n"
	}()
	// 实时发送消息到客户端
	for message := range messageChan {
		if _, err := ctx.Write([]byte(message)); err != nil {
			slog.Error("Failed to write SSE message", "error", err)
			break
		}
		// 强制刷新缓冲区
		if flusher, ok := ctx.Response().BodyWriter().(interface{ Flush() error }); ok {
			if err := flusher.Flush(); err != nil {
				slog.Error("Failed to flush response", "error", err)
				break
			}
		}
	}

	return nil
}

// Handler_DeployNomadJob 执行nomad一次性任务dispatch
func (n *NomadHandler) Handler_DeployNomadJob(ctx fiber.Ctx) error {
	n.setSSEHeaders(ctx)
	serverIDs := ctx.Query("server_ids")
	keys := strings.Split(serverIDs, ",")
	if len(keys) == 0 {
		return pkg.NewAppResponse(ctx, fiber.StatusBadRequest, 1, "server_ids is required", "", nil)
	}
	if len(keys) > 200 {
		return pkg.NewAppResponse(ctx, fiber.StatusBadRequest, 1, "server_ids is too many", "", nil)
	}
	messageChan := make(chan string, 200)
	results := make(chan task.JobResult, len(keys))
	go func() {
		//defer writer.Close()
		defer close(messageChan)
		contents := make([]map[string]string, 0)
		kv := n.Consul.KV()
		// 获取配置内容
		for _, key := range keys {
			content := make(map[string]string)
			pair, _, err := kv.Get(tools.AddNamespace(key, n.Ns), nil)
			if err != nil {
				messageChan <- fmt.Sprintf("data: [X] search config file failed. id: %s\n\n", key)
				continue
			}
			if pair == nil {
				messageChan <- fmt.Sprintf("data: [X] config file not found. id: %s\n\n", key)
				continue
			}
			content[key] = strings.TrimRight(string(pair.Value), "\r")
			contents = append(contents, content)
		}
		if len(contents) == 0 {
			messageChan <- "data: [X] all config files were not found! check!check again!\n\n"
			return
		}
		var wg sync.WaitGroup
		for _, content := range contents {
			for k, v := range content {
				job, err := n.Nomad.Jobs().ParseHCL(v, true)
				if err != nil {
					messageChan <- fmt.Sprintf("data: [X] parse job hcl config file failed. id: %s\n\n", k)
					continue
				}
				wg.Add(1)
				go func(j *nomadapi.Job, serverID string) {
					defer wg.Done()
					_, _, err := n.Nomad.Jobs().Register(j, nil)
					if err != nil {
						messageChan <- fmt.Sprintf("data: [X] register dispatch job failed. key: %s job: %s, error: %v\n\n", serverID, *j.ID, err)
						return
					}
					meta := make(map[string]string)
					meta["EXEC_TIME"] = time.Now().String()
					res, _, err := n.Nomad.Jobs().Dispatch(*j.ID, meta, []byte(time.Now().String()), "", nil)
					if err != nil {
						messageChan <- fmt.Sprintf("data: [X] deploy job failed. key: %s job: %s, error: %v\n\n", serverID, *j.ID, err)
						waitEvalCompletion(n.Nomad, serverID, res.EvalID, 120*time.Second, messageChan, results)
						return
					}
					waitEvalCompletion(n.Nomad, serverID, res.EvalID, 120*time.Second, messageChan, results)
					messageChan <- fmt.Sprintf("data: [√] deploy job success. key: %s job: %s, evalID: %s\n\n", serverID, *j.ID, res.EvalID)
				}(job, k)
			}
		}
		// 等待所有异步操作完成
		wg.Wait()
		close(results)
		messageChan <- "data: [√] All operations completed\n\n"
		messageChan <- fmt.Sprintln(tools.FormatLine("Job [result]", "*"))
		headers := []string{"EvalID", "JobID", "Type", "Status", "TriggeredBy"}
		drawTable(results, headers, messageChan)
	}()
	//return ctx.SendStream(reader)
	// 实时发送消息到客户端
	for message := range messageChan {
		if _, err := ctx.Write([]byte(message)); err != nil {
			slog.Error("Failed to write SSE message", "error", err)
			break
		}
		if flusher, ok := ctx.Response().BodyWriter().(interface{ Flush() error }); ok {
			if err := flusher.Flush(); err != nil {
				slog.Error("Failed to flush response", "error", err)
				break
			}
		}
	}
	return nil
}

// Handler_PurgeNomadJob 清除nomad job
func (n *NomadHandler) Handler_PurgeNomadJob(ctx fiber.Ctx) error {
	ids := ctx.Query("job_ids")
	if ids == "" {
		return pkg.NewAppResponse(ctx, fiber.StatusBadRequest, 1, "job_ids is required", "", nil)
	}
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
	return pkg.NewAppResponse(ctx, fiber.StatusOK, 0, "success", "", nil)
}

// waitEvalCompletion 轮询eval状态
func waitEvalCompletion(client *nomadapi.Client, serverID, evalID string, timeout time.Duration, messageChan chan string, ch chan task.JobResult) {
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
			messageChan <- fmt.Sprintf("data: timeout waiting for eval %s to complete,serverID: %s\n\n", evalID, serverID)
		case <-ticker.C:
			eval, _, err := client.Evaluations().Info(evalID, nil)
			if err != nil {
				messageChan <- fmt.Sprintf("data: failed to get job eval info. serverID: %s, error: %v\n\n", serverID, err)
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
func drawTable(ch chan task.JobResult, headers []string, messageChan chan string) {
	var data [][]string
	for result := range ch {
		item := []string{result.EvalID, result.JobID, result.Type, result.Status, result.TriggeredBy}
		data = append(data, item)
	}
	//writeTable(writer, headers, data)
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range data {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	tw.Flush()
	// SSE每行都加data:，否则浏览器不会显示
	for _, line := range strings.Split(buf.String(), "\n") {
		if strings.TrimSpace(line) != "" {
			messageChan <- "data: " + line + "\n"
		}
	}
	messageChan <- "\n"
}

// setSSEHeaders 设置SSE响应头
func (n *NomadHandler) setSSEHeaders(ctx fiber.Ctx) {
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")
	ctx.Set("Access-Control-Allow-Origin", "*")
	ctx.Set("Access-Control-Allow-Headers", "Cache-Control")
}

// updateGameStatus 更新游戏状态
func (n *NomadHandler) updateGameStatus(serverID string, status int) error {
	return config.DB.Exec("UPDATE games set status = ? where server_id = ?;", status, serverID).Error
}

// recordFailedJob 记录失败的任务
func (n *NomadHandler) recordFailedJob(mu *sync.Mutex, failCount *int, failedJobs *[]string, key string) {
	mu.Lock()
	defer mu.Unlock()
	*failCount++
	*failedJobs = append(*failedJobs, key)
}

// recordSuccessJob 记录成功的任务
func (n *NomadHandler) recordSuccessJob(mu *sync.Mutex, successCount *int, successJobs *[]string, key string) {
	mu.Lock()
	defer mu.Unlock()
	*successCount++
	*successJobs = append(*successJobs, key)
}
