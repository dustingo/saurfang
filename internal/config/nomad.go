package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	nomadapi "github.com/hashicorp/nomad/api"
)

const (
	NOMAD_HEALTH = "/v1/agent/health"
)

var (
	NomadCli     *nomadapi.Client
	nomadManager *NomadClient // 保存管理器实例，用于自动更新全局变量
)

type NomadClient struct {
	addresses        []string
	currentIndex     int
	maxRetries       int
	retryInterval    time.Duration
	ctx              context.Context
	cancelFunc       context.CancelFunc
	client           *nomadapi.Client
	clientLock       sync.RWMutex
	updateGlobalFunc func(*nomadapi.Client) // 用于更新全局变量的回调函数
}

// NewNomadClient 构建nomad客户端
func NewNomadClient(addresses []string, retryInterval time.Duration, maxRetries int) (*NomadClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	n := &NomadClient{
		addresses:     addresses,
		currentIndex:  0,
		maxRetries:    maxRetries,
		retryInterval: retryInterval,
		ctx:           ctx,
		cancelFunc:    cancel,
		client:        nil,
		clientLock:    sync.RWMutex{},
	}
	if err := n.connect(); err != nil {
		return nil, err
	}
	go n.nomadHealthCheck()
	return n, nil
}

// connect 尝试连接 Nomad 集群，如果当前节点失败则尝试下一个节点
func (n *NomadClient) connect() error {
	if len(n.addresses) == 0 {
		return errors.New("no nomad addresses configured")
	}
	var lastErr error
	// 从当前索引开始尝试所有节点
	numAddresses := len(n.addresses)
	for i := 0; i < numAddresses; i++ {
		idx := (n.currentIndex + i) % numAddresses
		address := strings.TrimSpace(n.addresses[idx])

		slog.Debug("Trying to connect to Nomad node", "address", address, "attempt", i+1, "total", numAddresses)

		config := &nomadapi.Config{
			Address: address,
		}
		client, err := nomadapi.NewClient(config)
		if err != nil {
			lastErr = fmt.Errorf("failed to create Nomad client for %s: %v", address, err)
			slog.Warn("Failed to create Nomad client", "address", address, "error", err)
			continue
		}

		// 验证连接是否可用
		if _, err := client.Status().Leader(); err != nil {
			lastErr = fmt.Errorf("failed to verify Nomad connection for %s: %v", address, err)
			slog.Warn("Nomad node connection verification failed", "address", address, "error", err)
			continue
		}

		// 连接成功，更新当前索引和客户端
		n.clientLock.Lock()
		n.client = client
		n.currentIndex = idx
		n.clientLock.Unlock()

		// 调用回调函数更新全局变量（关键：这样就能在重连后自动更新 NomadCli）
		if n.updateGlobalFunc != nil {
			n.updateGlobalFunc(client)
		}

		slog.Info("Successfully connected to Nomad", "address", address)
		return nil
	}
	return fmt.Errorf("failed to connect to any Nomad node: %v", lastErr)
}
func (n *NomadClient) reconnect() error {
	var lastErr error

	// connect() 会尝试所有节点，所以这里只需要重试 maxRetries 次
	// 每次从下一个节点开始尝试所有节点
	for i := 0; i < n.maxRetries; i++ {
		select {
		case <-n.ctx.Done():
			return errors.New("reconnect cancelled")
		default:
		}

		slog.Info("Attempting to reconnect to Nomad",
			"attempt", i+1,
			"maxRetries", n.maxRetries,
			"currentNode", n.GetCurrentAddress(),
		)

		// 尝试下一个节点（轮询）
		// 注意：connect() 会从这个索引开始尝试所有节点
		n.currentIndex = (n.currentIndex + 1) % len(n.addresses)

		if err := n.connect(); err != nil {
			lastErr = err
			slog.Warn("Reconnect attempt failed", "attempt", i+1, "error", err)

			// 如果这不是最后一次尝试，等待后再重试
			if i < n.maxRetries-1 {
				slog.Debug("Waiting before next retry", "interval", n.retryInterval)
				time.Sleep(n.retryInterval)
			}
			continue
		}

		// connect() 方法已经验证了连接，如果成功返回则说明连接可用
		slog.Info("Successfully reconnected to Nomad", "newNode", n.GetCurrentAddress())
		return nil
	}

	return fmt.Errorf("failed to reconnect after %d attempts (tried all nodes %d times): %v",
		n.maxRetries, n.maxRetries, lastErr)
}

// healthCheckLoop 健康检查循环
func (n *NomadClient) nomadHealthCheck() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !n.isNomadHealthy() {
				slog.Error("Nomad connection lost, attempting to reconnect...")
				if err := n.reconnect(); err != nil {
					slog.Error("Nomad reconnect failed", "error", err)
				}
			}
		case <-n.ctx.Done():
			return
		}
	}
}

// isHealthy 检查连接是否健康
func (n *NomadClient) isNomadHealthy() bool {
	n.clientLock.RLock()
	defer n.clientLock.RUnlock()
	if n.client == nil {
		return false
	}
	_, err := n.client.Status().Leader()
	return err == nil
}

// GetClient 获取 Nomad 客户端（线程安全）
func (n *NomadClient) GetClient() *nomadapi.Client {
	n.clientLock.RLock()
	defer n.clientLock.RUnlock()
	return n.client
}

// GetCurrentAddress 获取当前连接的地址
func (n *NomadClient) GetCurrentAddress() string {
	n.clientLock.RLock()
	defer n.clientLock.RUnlock()
	if n.currentIndex < len(n.addresses) {
		return n.addresses[n.currentIndex]
	}
	return ""
}

// Close 优雅关闭连接
func (n *NomadClient) Close() {
	if n.cancelFunc != nil {
		n.cancelFunc()
	}
	slog.Info("Nomad client closed")
}

// GetNomadManager 获取 Nomad 管理器实例
func GetNomadManager() *NomadClient {
	return nomadManager
}

func InitNomad() {
	nomadAddr := os.Getenv("NOMAD_HTTP_API_ADDR")
	if nomadAddr == "" {
		log.Fatalf("NOMAD_HTTP_API_ADDR environment variable is required")
	}

	// 支持多个地址，用逗号分隔（例如: "http://127.0.0.1:4646,http://127.0.0.1:4647,http://127.0.0.1:4648"）
	addresses := strings.Split(nomadAddr, ",")
	// 清理空白字符
	for i := range addresses {
		addresses[i] = strings.TrimSpace(addresses[i])
	}

	retryInterval, _ := strconv.Atoi(os.Getenv("NOMAD_RETRY_INTERVAL"))
	if retryInterval <= 0 {
		retryInterval = 5 // 默认5秒
	}

	maxRetries, _ := strconv.Atoi(os.Getenv("NOMAD_MAX_RETRIES"))
	if maxRetries <= 0 {
		maxRetries = 10 // 默认3次
	}

	client, err := NewNomadClient(addresses, time.Duration(retryInterval)*time.Second, maxRetries)
	if err != nil {
		log.Fatalf("NewNomadClient error: %v", err)
	}

	// 设置回调函数，用于在重连后自动更新全局变量
	// 这是关键！每次重连成功后，都会通过这个回调更新 NomadCli
	client.updateGlobalFunc = func(newClient *nomadapi.Client) {
		NomadCli = newClient
		slog.Info("Global NomadCli updated to new connection")
	}

	// 保存管理器实例和初始客户端
	nomadManager = client
	NomadCli = client.client

	slog.Info("Nomad initialized successfully",
		"addresses", addresses,
		"current", client.GetCurrentAddress(),
		"retryInterval", retryInterval,
		"maxRetries", maxRetries,
	)
}
