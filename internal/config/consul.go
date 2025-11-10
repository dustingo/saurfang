package config

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

var (
	ConsulCli     *consulapi.Client
	consulManager *ConsulClient // 保存管理器实例，用于自动更新全局变量
)

type ConsulClient struct {
	addresses        []string
	currentIndex     int
	config           *consulapi.Config
	client           *consulapi.Client
	clientLock       sync.RWMutex
	retryInterval    time.Duration
	maxRetries       int
	ctx              context.Context
	cancelFunc       context.CancelFunc
	updateGlobalFunc func(*consulapi.Client) // 用于更新全局变量的回调函数
}

// NewConsulClient new client with multiple addresses support
func NewConsulClient(addresses []string, retryInterval time.Duration, maxRetries int) (*ConsulClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &ConsulClient{
		addresses:     addresses,
		currentIndex:  0,
		retryInterval: retryInterval,
		maxRetries:    maxRetries,
		ctx:           ctx,
		cancelFunc:    cancel,
		config: &consulapi.Config{
			Token: os.Getenv("CONSUL_TOKEN"),
		},
	}
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect consul client error: %v", err)
	}
	go c.healthCheckLoop()
	return c, nil

}

// connect 尝试连接 Consul 集群，如果当前节点失败则尝试下一个节点
func (c *ConsulClient) connect() error {
	if len(c.addresses) == 0 {
		return fmt.Errorf("no Consul addresses configured")
	}

	var lastErr error
	// 从当前索引开始尝试所有节点
	for i := 0; i < len(c.addresses); i++ {
		idx := (c.currentIndex + i) % len(c.addresses)
		address := strings.TrimSpace(c.addresses[idx])

		// 创建配置，继承全局配置
		config := &consulapi.Config{
			Address:    address,
			Token:      c.config.Token,
			Scheme:     c.config.Scheme,
			Datacenter: c.config.Datacenter,
		}

		client, err := consulapi.NewClient(config)
		if err != nil {
			lastErr = fmt.Errorf("failed to create Consul client for %s: %v", address, err)
			slog.Warn("Failed to connect to Consul node", "address", address, "error", err)
			continue
		}

		// 验证连接是否可用
		if _, err := client.Status().Leader(); err != nil {
			lastErr = fmt.Errorf("failed to verify Consul connection for %s: %v", address, err)
			slog.Warn("Consul node connection verification failed", "address", address, "error", err)
			continue
		}

		// 连接成功，更新当前索引和客户端
		c.clientLock.Lock()
		c.client = client
		c.config = config
		c.currentIndex = idx
		c.clientLock.Unlock()

		// 调用回调函数更新全局变量（关键：这样就能在重连后自动更新 ConsulCli）
		if c.updateGlobalFunc != nil {
			c.updateGlobalFunc(client)
		}

		slog.Info("Successfully connected to Consul", "address", address)
		return nil
	}

	return fmt.Errorf("failed to connect to any Consul node: %v", lastErr)
}

// reconnect 重连方法，尝试所有节点
func (c *ConsulClient) reconnect() error {
	var lastErr error

	for i := 0; i < c.maxRetries; i++ {
		select {
		case <-c.ctx.Done():
			return fmt.Errorf("reconnect cancelled")
		default:
		}

		slog.Info("Attempting to reconnect to Consul", "attempt", i+1, "maxRetries", c.maxRetries)

		// 尝试下一个节点（轮询）
		c.currentIndex = (c.currentIndex + 1) % len(c.addresses)

		if err := c.connect(); err != nil {
			lastErr = err
			time.Sleep(c.retryInterval)
			continue
		}

		// connect() 方法已经验证了连接，如果成功返回则说明连接可用
		slog.Info("Successfully reconnected to Consul")
		return nil
	}

	return fmt.Errorf("failed to reconnect after %d attempts: %v", c.maxRetries, lastErr)
}

// healthCheckLoop 健康检查循环
func (c *ConsulClient) healthCheckLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !c.isHealthy() {
				slog.Error("Consul connection lost, attempting to reconnect...")
				if err := c.reconnect(); err != nil {
					slog.Error("Reconnect failed", "error", err)
				}
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// isHealthy 检查连接是否健康
func (c *ConsulClient) isHealthy() bool {
	c.clientLock.RLock()
	defer c.clientLock.RUnlock()

	if c.client == nil {
		return false
	}

	_, err := c.client.Status().Leader()
	return err == nil
}

// GetClient 获取 Consul 客户端（线程安全）
func (c *ConsulClient) GetClient() *consulapi.Client {
	c.clientLock.RLock()
	defer c.clientLock.RUnlock()
	return c.client
}

// GetCurrentAddress 获取当前连接的地址
func (c *ConsulClient) GetCurrentAddress() string {
	c.clientLock.RLock()
	defer c.clientLock.RUnlock()
	if c.currentIndex < len(c.addresses) {
		return c.addresses[c.currentIndex]
	}
	return ""
}

// Close 优雅关闭连接
func (c *ConsulClient) Close() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	slog.Info("Consul client closed")
}

// GetConsulManager 获取 Consul 管理器实例
func GetConsulManager() *ConsulClient {
	return consulManager
}
func InitConsul() {
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr == "" {
		log.Fatalf("CONSUL_ADDR environment variable is required")
	}

	// 支持多个地址，用逗号分隔（例如: "127.0.0.1:8500,127.0.0.1:8501,127.0.0.1:8502"）
	addresses := strings.Split(consulAddr, ",")
	// 清理空白字符
	for i := range addresses {
		addresses[i] = strings.TrimSpace(addresses[i])
	}

	retryInterval, _ := strconv.Atoi(os.Getenv("CONSUL_RETRY_INTERVAL"))
	if retryInterval <= 0 {
		retryInterval = 5 // 默认5秒
	}

	maxRetries, _ := strconv.Atoi(os.Getenv("CONSUL_MAX_RETRIES"))
	if maxRetries <= 0 {
		maxRetries = 3 // 默认3次
	}

	client, err := NewConsulClient(addresses, time.Duration(retryInterval)*time.Second, maxRetries)
	if err != nil {
		log.Fatalf("NewConsulClient error: %v", err)
	}

	// 配置额外选项（可选）
	if scheme := os.Getenv("CONSUL_SCHEME"); scheme != "" {
		client.config.Scheme = scheme // http 或 https
	}
	if datacenter := os.Getenv("CONSUL_DATACENTER"); datacenter != "" {
		client.config.Datacenter = datacenter
	}

	// 设置回调函数，用于在重连后自动更新全局变量
	// 这是关键！每次重连成功后，都会通过这个回调更新 ConsulCli
	client.updateGlobalFunc = func(newClient *consulapi.Client) {
		ConsulCli = newClient
		slog.Info("Global ConsulCli updated to new connection")
	}

	// 保存管理器实例和初始客户端
	consulManager = client
	ConsulCli = client.client

	slog.Info("Consul initialized successfully",
		"addresses", addresses,
		"current", client.GetCurrentAddress(),
		"scheme", client.config.Scheme,
		"datacenter", client.config.Datacenter,
		"retryInterval", retryInterval,
		"maxRetries", maxRetries,
	)
}
