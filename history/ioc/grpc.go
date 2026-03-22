package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	grpc2 "webook/history/grpc"
	"webook/pkg/grpcx"
	"webook/pkg/logger"
)

func InitGRPCServer(historyServer *grpc2.HistoryServiceServer, l logger.LoggerV1) *grpcx.Server {
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
	historyServer.Register(server)

	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "history",
		L:         l,
	}
}
