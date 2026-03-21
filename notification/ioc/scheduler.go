package ioc

import (
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"

	"webook/notification/domain"
	"webook/notification/repository"
	"webook/notification/scheduler"
	"webook/notification/service"
	"webook/notification/service/channel"
	"webook/pkg/cronjobx"
	"webook/pkg/logger"
)

func InitETCDClient() *clientv3.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints: cfg.Addrs,
	})
	if err != nil {
		panic(err)
	}
	return client
}

func InitCheckBackScheduler(
	txRepo repository.TransactionRepository,
	svc service.NotificationService,
	etcdClient *clientv3.Client,
	l logger.LoggerV1,
) *scheduler.CheckBackScheduler {
	return scheduler.NewCheckBackScheduler(txRepo, svc, etcdClient, l)
}

func InitScheduledSendJob(
	repo repository.NotificationRepository,
	senders map[domain.Channel]channel.Sender,
	l logger.LoggerV1,
) *scheduler.ScheduledSendJob {
	return scheduler.NewScheduledSendJob(repo, senders, l)
}

func InitCronJobs(l logger.LoggerV1, checkBack *scheduler.CheckBackScheduler, scheduledSend *scheduler.ScheduledSendJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cdb := cronjobx.NewCronJobBuilder(l)
	// 每 10 秒执行一次事务回查
	_, err := res.AddJob("*/10 * * * * ?", cdb.Build(checkBack))
	if err != nil {
		panic(err)
	}
	// 每 5 秒扫描一次到期的延迟发送通知
	_, err = res.AddJob("*/5 * * * * ?", cdb.Build(scheduledSend))
	if err != nil {
		panic(err)
	}
	return res
}
