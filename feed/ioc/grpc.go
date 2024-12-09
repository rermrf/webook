package ioc

import (
	"github.com/spf13/viper"
	igrpc "webook/feed/grpc"
	"webook/pkg/grpcx"

	"google.golang.org/grpc"
	"webook/pkg/logger"
)

func InitGRPCServer(feedSvc *igrpc.FeedEventGrpcServer, l logger.LoggerV1) *grpcx.Server {
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
	feedSvc.Register(server)

	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "feed",
		L:         l,
	}
}
