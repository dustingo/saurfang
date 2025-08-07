package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type DBManager struct {
	db          *gorm.DB
	dsn         string
	maxRetries  int
	retryDelay  time.Duration
	checkPeriod time.Duration
}

func NewDBManager(dsn string) *DBManager {
	return &DBManager{
		dsn:         dsn,
		maxRetries:  5,
		retryDelay:  time.Second * 5,
		checkPeriod: time.Minute * 5,
	}
}
func (c *DBManager) Connect() error {
	var err error
	for i := 0; i < c.maxRetries; i++ {
		c.db, err = gorm.Open(mysql.Open(c.dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err == nil {
			// 设置连接池配置
			sqlDB, err := c.db.DB()
			if err == nil {
				sqlDB.SetMaxIdleConns(10)
				sqlDB.SetMaxOpenConns(100)
				sqlDB.SetConnMaxLifetime(time.Hour)
			}
			go c.periodicHealthCheck()
			return nil
		}
		time.Sleep(c.retryDelay)
	}
	return fmt.Errorf("failed to connect after %d attempts: %w", c.maxRetries, err)
}
func (c *DBManager) periodicHealthCheck() {
	ticker := time.NewTicker(c.checkPeriod)
	for range ticker.C {
		if err := c.db.Exec("SELECT 1").Error; err != nil {
			slog.Error("database connection lost", "error", err)
			if err := c.Connect(); err != nil {
				slog.Error("failed to reconnect", "error", err)
			} else {
				slog.Info("successfully reconnected to the database")
			}
		}
	}
}
func (c *DBManager) DB() *gorm.DB {
	return c.db
}

func InitMySQL() {
	User := os.Getenv("MYSQL_USER")
	PASSWORD := os.Getenv("MYSQL_PASSWORD")
	HOST := os.Getenv("MYSQL_HOST")
	PORT := os.Getenv("MYSQL_PORT")
	DATABASE := os.Getenv("MYSQL_DB")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci", User,
		PASSWORD, HOST, PORT, DATABASE)
	mysqlManager := NewDBManager(dsn)
	if err := mysqlManager.Connect(); err != nil {
		slog.Error("failed to connect to the database", "error", err)
		os.Exit(-1)
	}
	DB = mysqlManager.DB()
}
