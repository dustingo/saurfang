// Package config 初始化数据库redis缓存链接
package config

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var NtfyClient *redis.Client

type NtfyRedisManager struct {
	config      *redis.Options
	client      *redis.Client
	ctx         context.Context
	mu          sync.Mutex
	stopChan    chan struct{}
	isConnected bool
}

// NewNtfyRedisManager 创建Redis管理器
func NewNtfyRedisManager(addr, password string, db int) *NtfyRedisManager {

	config := &redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	manager := &NtfyRedisManager{
		config:   config,
		ctx:      context.Background(),
		stopChan: make(chan struct{}),
	}

	// 初始化连接
	manager.connect()

	// 启动健康检查
	go manager.healthCheck()

	return manager
}

// connect 内部连接方法
func (n *NtfyRedisManager) connect() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.client != nil {
		n.client.Close()
	}

	n.client = redis.NewClient(n.config)
	if err := n.client.Ping(n.ctx).Err(); err != nil {
		n.isConnected = false
		slog.Error("Redis connection failed", "error", err)
	} else {
		n.isConnected = true
		slog.Info("Connected to Redis successfully")
	}
}

// healthCheck 健康检查
func (m *NtfyRedisManager) healthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			if m.client == nil {
				m.mu.Unlock()
				continue
			}

			if err := m.client.Ping(m.ctx).Err(); err != nil {
				m.isConnected = false
				slog.Error("Redis connection lost, attempting to reconnect...")
				m.connect()
			} else if !m.isConnected {
				m.isConnected = true
				slog.Info("Redis connection restored")
			}
			m.mu.Unlock()
		case <-m.stopChan:
			return
		}
	}
}

// GetClient 获取Redis客户端
func (m *NtfyRedisManager) GetClient() (*redis.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isConnected || m.client == nil {
		return nil, errors.New("redis is not connected")
	}
	return m.client, nil
}

// Close 关闭连接和健康检查
func (m *NtfyRedisManager) Close() error {
	close(m.stopChan)
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		return m.client.Close()
	}
	return nil
}
func InitNtfy() *redis.Client {
	db, err := strconv.Atoi(os.Getenv("REDIS_PUB_SUB_DB"))
	if err != nil {
		slog.Error("notify  db parse error:", "error", err)
		os.Exit(-1)
	}
	ntfyMgr := NewRedisManager(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PUB_SUB_DB"), db)
	NtfyClient, err = ntfyMgr.GetClient()
	if err != nil {
		slog.Error("notify init error:", "error", err)
		os.Exit(-1)
	}
	return NtfyClient
}
