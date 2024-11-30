package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"webook/pkg/saramax"
	"webook/search/events"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	// 使用 hash 保证同一个 biz 发送到同一个 topic
	// 如果 要新增分区，怎么保证消息的顺序性？
	// 在原本分区没有消息挤压的前提下，让新分区睡眠一小段时间，等待之前的消息消费完
	saramaCfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
func NewConsumers(articleConsumer *events.ArticleConsumer, userConsumer *events.UserConsumer) []saramax.Consumer {
	return []saramax.Consumer{
		articleConsumer,
		userConsumer,
	}
}
