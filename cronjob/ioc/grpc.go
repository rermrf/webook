package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	igrpc "webook/cronjob/grpc"
	"webook/pkg/grpcx"
)

func InitGRPCxServer(cronJobServer *igrpc.CronJobServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	cronJobServer.Register(server)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
