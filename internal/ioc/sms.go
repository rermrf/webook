package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	smsv1 "webook/api/proto/gen/sms/v1"
)

func InitSMSGRPCClient(client *etcdv3.Client) smsv1.SMSServiceClient {
	type config struct {
		Secure bool `yaml:"secure"`
	}
	var cfg config
	err := viper.UnmarshalKey("grpc.client.sms", &cfg)
	if err != nil {
		panic(err)
	}

	bd, err := resolver.NewBuilder(client)
	if err != nil {
		panic(err)
	}

	opts := []grpc.DialOption{grpc.WithResolvers(bd)}
	if cfg.Secure {
		// 加载证书之类的东西
		// 启用 HTTPS
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient("etcd:///service/sms", opts...)
	if err != nil {
		panic(err)
	}
	return smsv1.NewSMSServiceClient(cc)
}
