package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"webook/notification/events"
	"webook/pkg/saramax"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(cfg.Addrs, scfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSyncProducer(client sarama.Client) sarama.SyncProducer {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return producer
}

func NewConsumers(
	notificationConsumer *events.NotificationEventConsumer,
	likeConsumer *events.LikeEventConsumer,
	collectConsumer *events.CollectEventConsumer,
	commentConsumer *events.CommentEventConsumer,
	followConsumer *events.FollowEventConsumer,
) []saramax.Consumer {
	return []saramax.Consumer{
		notificationConsumer,
		likeConsumer,
		collectConsumer,
		commentConsumer,
		followConsumer,
	}
}
