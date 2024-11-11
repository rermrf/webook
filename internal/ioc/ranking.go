package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
	rankingv1 "webook/api/proto/gen/ranking/v1"
	"webook/internal/job"
	"webook/pkg/cronjobx"
	"webook/pkg/logger"
)

func InitRankingJob(svc rankingv1.RankingServiceClient, client *rlock.Client, l logger.LoggerV1) *job.RankingJob {
	return job.NewRankingJob(svc, time.Second*30, client, l)
}

func InitJob(l logger.LoggerV1, ranking *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cdb := cronjobx.NewCronJobBuilder(l)
	// 每三分钟一次
	_, err := res.AddJob("0 */3 * * * ?", cdb.Build(ranking))
	if err != nil {
		panic(err)
	}
	return res
}

func InitRankingGRPCClient(client *etcdv3.Client) rankingv1.RankingServiceClient {
	type Config struct {
		Secure bool `json:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.ranking", &cfg)
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
	cc, err := grpc.NewClient("etcd:///service/ranking", opts...)
	if err != nil {
		panic(err)
	}
	return rankingv1.NewRankingServiceClient(cc)
}
