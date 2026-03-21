package ioc

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	notificationv2 "webook/api/proto/gen/notification/v2"
	"webook/bff/handler/sse"
	"webook/pkg/logger"
)

func InitNotificationGRPCClient(client *etcdv3.Client) notificationv2.NotificationServiceClient {
	type Config struct {
		Secure bool `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.notification", &cfg)
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

	cc, err := grpc.NewClient("etcd:///service/notification", opts...)
	if err != nil {
		panic(err)
	}
	return notificationv2.NewNotificationServiceClient(cc)
}

// InitSSEHub 初始化 SSE Hub
func InitSSEHub(redisClient redis.Cmdable, l logger.LoggerV1) *sse.Hub {
	hub := sse.NewHub(redisClient, l)
	// 启动 Hub（在后台运行）
	go hub.Run(context.Background())
	return hub
}
