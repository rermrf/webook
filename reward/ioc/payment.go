package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pmtv1 "webook/api/proto/gen/payment/v1"
)

func InitPaymentGRPCClientV1(client *etcdv3.Client) pmtv1.WechatPaymentServiceClient {
	type Config struct {
		Target string `yaml:"target"`
		Secure bool   `json:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.payment", &cfg)
	if err != nil {
		panic(err)
	}

	bd, err := resolver.NewBuilder(client)
	if err != nil {
		panic(err)
	}

	opts := []grpc.DialOption{
		grpc.WithResolvers(bd),
		// 使用轮训负载均衡算法
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"round_robin":{}}]}`),
		//grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"weighted_round_robin": {}}]}`),
	}
	if cfg.Secure {
		// 加载证书之类的东西
		// 启用 HTTPS
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient(cfg.Target, opts...)
	if err != nil {
		panic(err)
	}
	return pmtv1.NewWechatPaymentServiceClient(cc)
}
