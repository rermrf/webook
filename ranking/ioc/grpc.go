package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"webook/pkg/grpcx"
	igrpc "webook/ranking/grpc"
)

func InitGRPCServer(rankServer *igrpc.RankingServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	rankServer.Register(server)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
