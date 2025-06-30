package config

import (
	"os"
	"saurfang/internal/models/db"
	"strconv"
)

var SynqConfig db.AsynqConfig

func InitSynq() {
	SynqConfig.Addr = os.Getenv("REDIS_HOST")
	SynqConfig.Password = os.Getenv("REDIS_PASSWORD")
	SynqConfig.DB, _ = strconv.Atoi(os.Getenv("REDIS_DB"))
	SynqConfig.Queue = os.Getenv("REDIS_QUEUE")
	SynqConfig.Concurrency, _ = strconv.Atoi(os.Getenv("REDIS_CONCURRENCY"))
	SynqConfig.SyncInterval, _ = strconv.Atoi(os.Getenv("REDIS_SYNC_INTERVAL"))
	SynqConfig.Location = os.Getenv("REDIS_LOCATION")
}
