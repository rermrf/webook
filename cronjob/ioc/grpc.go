package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	igrpc "webook/cronjob/grpc"
	"webook/pkg/grpcx"
	"webook/pkg/logger"
)

func InitGRPCxServer(cronJobServer *igrpc.CronJobServiceServer, l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port      int      `yaml:"port"`
		EtcdAddrs []string `yaml:"etcdAddrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	cronJobServer.Register(server)

	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "cornjob",
		L:         l,
	}
}
