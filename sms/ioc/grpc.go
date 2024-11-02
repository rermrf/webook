package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"webook/pkg/grpcx"
	igrpc "webook/sms/grpc"
)

func InitGRPCServer(smsServer *igrpc.SMSGRPCServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	smsServer.Register(server)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
