// Asynq配置
package db

type AsynqConfig struct {
	Addr         string `yaml:"addr"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	Queue        string `yaml:"queue"`
	Concurrency  int    `yaml:"concurrency"`
	SyncInterval int    `yaml:"syncInterval"`
	Location     string `yaml:"location"`
}
