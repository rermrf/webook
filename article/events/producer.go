package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProducerReadEvent(ctx context.Context, evt ReadEvent) error
	ProducerReadEventV1(ctx context.Context, evt ReadEventV1)
	ProduceSyncEvent(ctx context.Context, evt SyncArticleEvent) error
}

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(pc sarama.SyncProducer) Producer {
	return &KafkaProducer{
		producer: pc,
	}
}

// ProducerReadEvent 如果有复杂的重试逻辑，就用装饰器
// 如果你认为你的重试逻辑很简单，你就放这里
func (k *KafkaProducer) ProducerReadEvent(ctx context.Context, evt ReadEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "read_article",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func (k *KafkaProducer) ProduceSyncEvent(ctx context.Context, evt SyncArticleEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "sync_article_event",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func (k *KafkaProducer) ProducerReadEventV1(ctx context.Context, evt ReadEventV1) {
	//TODO implement me
	panic("implement me")
}

type ReadEvent struct {
	Uid int64
	Aid int64
}

type ReadEventV1 struct {
	Uids []int64
	Aids []int64
}

type SyncArticleEvent struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Status  int32  `json:"status"`
	Content string `json:"content"`
}
