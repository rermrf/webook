package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	smsv1 "webook/api/proto/gen/sms/v1"
)

func InitEtcd() *etcdv3.Client {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("etcd", &cfg); err != nil {
		panic(err)
	}
	cli, err := etcdv3.NewFromURL(cfg.Addr)
	if err != nil {
		panic(err)
	}
	return cli
}

func InitSmsGRPCClient(client *etcdv3.Client) smsv1.SMSServiceClient {
	bd, err := resolver.NewBuilder(client)
	if err != nil {
		panic(err)
	}

	opts := []grpc.DialOption{grpc.WithResolvers(bd), grpc.WithTransportCredentials(insecure.NewCredentials())}
	cc, err := grpc.NewClient("etcd:///service/sms", opts...)
	if err != nil {
		panic(err)
	}
	return smsv1.NewSMSServiceClient(cc)
}
