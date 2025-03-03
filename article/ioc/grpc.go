package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	igrpc "webook/article/grpc"
	"webook/pkg/grpcx"
	"webook/pkg/logger"
)

func InitGRPCServer(artServer *igrpc.ArticleGRPCServer, l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port      int      `yaml:"port"`
		EtcdAddrs []string `yaml:"etcdAddrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	// 在这里添加服务端拦截器
	server := grpc.NewServer(grpc.ChainUnaryInterceptor())
	artServer.Register(server)

	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "article",
		L:         l,
	}
}
