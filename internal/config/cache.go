// Package config 初始化数据库redis缓存链接
package config

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var CahceClient *redis.Client

type RedisManager struct {
	config      *redis.Options
	client      *redis.Client
	ctx         context.Context
	mu          sync.Mutex
	stopChan    chan struct{}
	isConnected bool
}

// NewRedisManager 创建Redis管理器
func NewRedisManager(addr, password string, db int) *RedisManager {
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

	manager := &RedisManager{
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
func (m *RedisManager) connect() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		m.client.Close()
	}

	m.client = redis.NewClient(m.config)
	if err := m.client.Ping(m.ctx).Err(); err != nil {
		m.isConnected = false
		log.Printf("Redis connection failed: %v", err)
	} else {
		m.isConnected = true
		log.Println("Connected to Redis successfully")
	}
}

// healthCheck 健康检查
func (m *RedisManager) healthCheck() {
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
				log.Printf("Redis connection lost, attempting to reconnect...")
				m.connect()
			} else if !m.isConnected {
				m.isConnected = true
				log.Println("Redis connection restored")
			}
			m.mu.Unlock()
		case <-m.stopChan:
			return
		}
	}
}

// GetClient 获取Redis客户端
func (m *RedisManager) GetClient() (*redis.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isConnected || m.client == nil {
		return nil, errors.New("Redis is not connected")
	}

	return m.client, nil
}

// Close 关闭连接和健康检查
func (m *RedisManager) Close() error {
	close(m.stopChan)
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		return m.client.Close()
	}
	return nil
}
func InitCache() *redis.Client {
	db, err := strconv.Atoi(os.Getenv("REDIS_CACHE_DB"))
	if err != nil {
		log.Fatalln("Redis cache db parse error:", err)
	}
	cacheMgr := NewRedisManager(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PASSWORD"), db)
	CahceClient, err = cacheMgr.GetClient()
	if err != nil {
		log.Fatalln("Redis cache init error:", err)
	}
	return CahceClient
}
