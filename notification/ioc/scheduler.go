package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"

	"webook/notification/repository"
	"webook/notification/scheduler"
	"webook/notification/service"
	"webook/pkg/logger"
)

func InitETCDClient() *clientv3.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints: cfg.Addrs,
	})
	if err != nil {
		panic(err)
	}
	return client
}

func InitCheckBackScheduler(
	txRepo repository.TransactionRepository,
	svc service.NotificationService,
	etcdClient *clientv3.Client,
	l logger.LoggerV1,
) *scheduler.CheckBackScheduler {
	return scheduler.NewCheckBackScheduler(txRepo, svc, etcdClient, l)
}
