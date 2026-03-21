package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	grpc2 "webook/im/grpc"
	"webook/pkg/grpcx"
	"webook/pkg/logger"
)

func InitGRPCServer(imServer *grpc2.IMServiceServer, l logger.LoggerV1) *grpcx.Server {
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
	imServer.Register(server)

	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "im",
		L:         l,
	}
}
