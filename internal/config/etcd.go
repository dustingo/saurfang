package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var Etcd *clientv3.Client

type EtcdManager struct {
	db          *clientv3.Client
	endPoints   []string
	username    string
	password    string
	maxRetries  int
	retryDelay  time.Duration
	checkPeriod time.Duration
}

func NewEtcdanager(endPoints []string, username, password string) *EtcdManager {
	return &EtcdManager{
		endPoints:   endPoints,
		username:    username,
		password:    password,
		maxRetries:  5,
		retryDelay:  time.Second * 5,
		checkPeriod: time.Minute * 5,
	}
}

func (m *EtcdManager) Connect() error {
	var err error
	for i := 0; i < m.maxRetries; i++ {
		m.db, err = clientv3.New(clientv3.Config{
			Endpoints:   m.endPoints,
			Username:    m.username,
			Password:    m.password,
			DialTimeout: 8 * time.Second,
		})
		if err == nil {
			go m.periodicHealthCheck()
			return nil
		}
		time.Sleep(m.retryDelay)
	}
	return fmt.Errorf("failed to connect etcd after %d attempts: %w", m.maxRetries, err)
}

func (m *EtcdManager) periodicHealthCheck() {
	ticker := time.NewTicker(m.checkPeriod)
	for range ticker.C {
		_, err := clientv3.New(clientv3.Config{
			Endpoints:   m.endPoints,
			Username:    m.username,
			Password:    m.password,
			DialTimeout: 8 * time.Second,
		})
		if err != nil {
			log.Printf("etcd connection lost: %v", err)
			if err := m.Connect(); err != nil {
				log.Printf("filed to reconnect: %v", err)
			} else {
				log.Println("successfully reconnected to the etcd")
			}
		}
	}
}

func (m *EtcdManager) DB() *clientv3.Client {
	return m.db
}

var etcdManager *EtcdManager

// GetDB returns the current database connection
func GetEtcd() *clientv3.Client {
	return etcdManager.DB()
}
func InitEtcd() {
	endPoint := os.Getenv("ETCD_ENDPOINT")
	username := os.Getenv("ETCD_USERNAME")
	password := os.Getenv("ETCD_PASSWORD")
	manager := NewEtcdanager(strings.Split(endPoint, ","), username, password)
	if err := manager.Connect(); err != nil {
		log.Println("etcd connection lost")
		os.Exit(1)
	}
	Etcd = manager.DB()
}
