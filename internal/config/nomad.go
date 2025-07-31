package config

import (
	nomadapi "github.com/hashicorp/nomad/api"
	"os"
)

// NewNomadClient 构建nomad客户端
func NewNomadClient() (*nomadapi.Client, error) {
	client, err := nomadapi.NewClient(&nomadapi.Config{
		Address: os.Getenv("NOMAD_HTTP_API_ADDR"),
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
