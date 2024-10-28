package startup

import (
	"github.com/IBM/sarama"
	"webook/internal/events"
	"webook/internal/events/article"
)

func InitKafka() sarama.Client {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9094"}, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

// NewConsumer 面临的问题依旧是所有的 Consumer 在这里注册一下
//func NewConsumer(c1 *article.InteractiveReadEventConsumer) []events.Consumer {
//	return []events.Consumer{c1}
//}

func NewConsumer(c1 *article.InteractiveReadBatchConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
