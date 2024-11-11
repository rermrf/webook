package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	igrpc "webook/oauth2/grpc"
	"webook/pkg/grpcx"
	"webook/pkg/logger"
)

func InitGRPCxServer(oauth2Server *igrpc.Oauth2ServiceServer, l logger.LoggerV1) *grpcx.Server {
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
	oauth2Server.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		EtcdAddrs: cfg.EtcdAddrs,
		Name:      "oauth2",
		L:         l,
	}
}
