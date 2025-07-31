package config

import (
	"context"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var ConsulCli *consulapi.Client

type ConsulClient struct {
	address       string
	config        *consulapi.Config
	client        *consulapi.Client
	clientLock    sync.RWMutex
	retryInterval time.Duration
	maxRetries    int
	ctx           context.Context
	cancelFunc    context.CancelFunc
}

// NewConsulClient new client
func NewConsulClient(address string, retryInterval time.Duration, maxRetries int) (*ConsulClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &ConsulClient{
		address:       address,
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
func (c *ConsulClient) connect() error {
	config := c.config
	client, err := consulapi.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Consul client: %v", err)
	}

	c.clientLock.Lock()
	defer c.clientLock.Unlock()
	c.client = client
	c.config = config
	return nil
}

// reconnect 重连方法
func (c *ConsulClient) reconnect() error {
	var lastErr error

	for i := 0; i < c.maxRetries; i++ {
		select {
		case <-c.ctx.Done():
			return fmt.Errorf("reconnect cancelled")
		default:
		}

		log.Printf("Attempting to reconnect to Consul (attempt %d/%d)", i+1, c.maxRetries)

		if err := c.connect(); err != nil {
			lastErr = err
			time.Sleep(c.retryInterval)
			continue
		}

		// 验证连接是否真的可用
		if _, err := c.client.Status().Leader(); err == nil {
			log.Println("Successfully reconnected to Consul")
			return nil
		}
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
				log.Println("Consul connection lost, attempting to reconnect...")
				if err := c.reconnect(); err != nil {
					log.Printf("Reconnect failed: %v", err)
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
func InitConsul() {
	consulAddr := os.Getenv("CONSUL_ADDR")
	retryInterval, _ := strconv.Atoi(os.Getenv("CONSUL_RETRY_INTERVAL"))
	maxRetries, _ := strconv.Atoi(os.Getenv("CONSUL_MAX_RETRIES"))
	client, err := NewConsulClient(consulAddr, time.Duration(retryInterval)*time.Second, maxRetries)
	if err != nil {
		log.Fatalf("NewConsulClient error: %v", err)
	}
	ConsulCli = client.client
}
