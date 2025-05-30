package db

import (
	"time"

	"gorm.io/gorm"
)

type DBManager struct {
	Db          *gorm.DB
	Dsn         string
	MaxRetries  int
	RetryDelay  time.Duration
	CheckPeriod time.Duration
}
