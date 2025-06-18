package gameservice

import (
	"context"
	"errors"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"saurfang/internal/models/serverconfig"
	"saurfang/internal/repository/base"
)

// ServerConfigService
type ServerConfigService struct {
	base.BaseGormRepository[serverconfig.Configs]
	Ns string
}

// NewServerConfigService
func NewServerConfigService(etcd *clientv3.Client, ns string) *ServerConfigService {
	return &ServerConfigService{
		BaseGormRepository: base.BaseGormRepository[serverconfig.Configs]{Etcd: etcd},
		Ns:                 ns,
	}
}

func (s *ServerConfigService) Service_CreateGameConfig(key, setting string) error {
	res, err := s.Etcd.Get(context.Background(), s.addNamespace(key))
	if err != nil {
		return err
	}
	if len(res.Kvs) > 0 {
		return errors.New("key already exists")
	}
	if _, err := s.Etcd.Put(context.Background(), s.addNamespace(key), setting); err != nil {
		return err
	}
	return nil
}
func (s *ServerConfigService) Service_DeleteGameConfig(key string) error {
	if _, err := s.Etcd.Delete(context.Background(), s.addNamespace(key)); err != nil {
		return err
	}
	return nil
}
func (s *ServerConfigService) Service_UpdateGameConfig(key, setting string) error {
	res, err := s.Etcd.Get(context.Background(), s.addNamespace(key))
	if err != nil {
		return err
	}
	if len(res.Kvs) > 0 {
		_, err := s.Etcd.Delete(context.Background(), s.addNamespace(key))
		if err != nil {
			return err
		}
		_, err = s.Etcd.Put(context.Background(), s.addNamespace(key), setting)
		if err != nil {
			return err
		}
	} else {
		return errors.New("key does not exist")
	}
	return nil
}

func (s *ServerConfigService) Service_ListGameConfig() ([]serverconfig.GameConfigDto, error) {
	var items []serverconfig.GameConfigDto
	var item serverconfig.GameConfigDto
	data, err := s.Etcd.Get(context.Background(), s.Ns, clientv3.WithPrefix())
	if err != nil {
		return items, err
	}
	for _, ev := range data.Kvs {
		item.Key = s.removeNamespace(string(ev.Key))
		item.Setting = string(ev.Value)
		items = append(items, item)
	}
	return items, nil
}

func (s *ServerConfigService) Service_ListGameConfigBykey(key string) (*serverconfig.GameConfigDto, error) {

	var item serverconfig.GameConfigDto
	data, err := s.Etcd.Get(context.Background(), s.addNamespace(key))
	if err != nil {
		return nil, err
	}
	if len(data.Kvs) == 0 {
		return nil, errors.New("key not found")
	}
	item.Key = string(data.Kvs[0].Key)
	item.Setting = string(data.Kvs[0].Value)
	return &item, nil
}
func (s *ServerConfigService) addNamespace(key string) string {
	return fmt.Sprintf("%s/%s", s.Ns, key)
}
func (s *ServerConfigService) removeNamespace(key string) string {
	return key[len(s.Ns)+1:]
}
