package ioc

import (
	"github.com/spf13/viper"
	grpc2 "webook/payment/grpc"
	"webook/pkg/grpcx"

	"google.golang.org/grpc"
	"webook/pkg/logger"
)

func InitGRPCServer(weSvc *grpc2.WechatServiceServer, l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port      int      `yaml:"port"`
		EtcdAddrs []string `yaml:"etcdAddrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer(grpc.ChainUnaryInterceptor())
	weSvc.Register(server)

	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "article",
		L:         l,
	}
}
