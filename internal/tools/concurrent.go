package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cast"
	"os"
	"os/exec"
	"saurfang/internal/config"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/models/task.go"
	"strconv"
	"sync"
)

func ConcurrentExec(data map[string]serverconfig.Configs, c fiber.Ctx, ops, serverid string) {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	// 同步任务的结果
	results := make(chan task.OpsStatus, cast.ToInt(os.Getenv("EXEC_CONCURRENCY")))
	var wg sync.WaitGroup
	concurrency := cast.ToInt(os.Getenv("EXEC_CONCURRENCY"))
	sem := make(chan struct{}, concurrency)
	for _, v := range data {
		wg.Add(1)
		sem <- struct{}{}
		go func(cnf *serverconfig.Configs, ops string) {
			defer wg.Done()
			defer func() { <-sem }()
			taskResult := task.OpsStatus{}
			shell := []string{fmt.Sprintf("-p%s", strconv.Itoa(cnf.Port)), fmt.Sprintf("%s@%s", cnf.User, cnf.IP), fmt.Sprintf("cd %s && ", cnf.ConfigDir)}
			switch ops {
			case "start":
				shell = append(shell, cnf.Start)
			case "stop":
				shell = append(shell, cnf.Stop)
			}
			cmd := exec.Command("ssh", shell...)
			var stdErr bytes.Buffer
			cmd.Stderr = &stdErr
			err := cmd.Run()
			if err != nil {
				taskResult.Status = "failure"
				taskResult.ErrorMessage = stdErr.String()
				taskResult.SvcName = cnf.SvcName
				taskResult.ServerID = cnf.ServerId
				results <- taskResult
				c.WriteString(fmt.Sprintf("%s ServerID:%s IP:%s SvcName:%s Failed|%s\n", ops, cnf.ServerId, cnf.IP, cnf.SvcName, stdErr.String()))
				if bw, ok := c.Response().BodyWriter().(*bufio.Writer); ok {
					bw.Flush()
				}
				return
			}
			taskResult.Status = "success"
			taskResult.ErrorMessage = stdErr.String()
			taskResult.SvcName = cnf.SvcName
			taskResult.ServerID = cnf.ServerId
			results <- taskResult
			c.WriteString(fmt.Sprintf("%s ServerID:%s IP:%s SvcName:%s Success\n", ops, cnf.ServerId, cnf.IP, cnf.SvcName))
			if bw, ok := c.Response().BodyWriter().(*bufio.Writer); ok {
				bw.Flush()
			}
		}(&v, ops)
	}
	wg.Wait()
	close(results)
	var total, success, failed int
	for result := range results {
		total++
		switch result.Status {
		case "success":
			success++
		case "failure":
			failed++
		}
	}
	c.WriteString(FormatLine("TASK [result]", "*"))
	c.WriteString(fmt.Sprintf("Total: %d\t  Success:%d\t  Failed:%d\n", total, success, failed))
	if bw, ok := c.Response().BodyWriter().(*bufio.Writer); ok {
		bw.Flush()
	}
	if ops == "start" {
		if failed == 0 {
			config.DB.Exec("update saurfang_games set status = 1 where server_id = ?;", serverid)

		}
	} else if ops == "stop" {
		if failed == 0 {
			config.DB.Exec("update saurfang_games set status = 0 where server_id = ?;", serverid)
		}
	}
}
