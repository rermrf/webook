package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

// Producer 通知事件生产者接口
type Producer interface {
	ProduceNotificationEvent(ctx context.Context, evt NotificationEvent) error
	ProduceLikeEvent(ctx context.Context, evt LikeEvent) error
	ProduceCommentEvent(ctx context.Context, evt CommentEvent) error
	ProduceFollowEvent(ctx context.Context, evt FollowEvent) error
}

type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaProducer(producer sarama.SyncProducer) Producer {
	return &SaramaProducer{
		producer: producer,
	}
}

func (p *SaramaProducer) ProduceNotificationEvent(ctx context.Context, evt NotificationEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicNotificationEvents,
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func (p *SaramaProducer) ProduceLikeEvent(ctx context.Context, evt LikeEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "like_events",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func (p *SaramaProducer) ProduceCommentEvent(ctx context.Context, evt CommentEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "comment_events",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func (p *SaramaProducer) ProduceFollowEvent(ctx context.Context, evt FollowEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "follow_events",
		Value: sarama.ByteEncoder(data),
	})
	return err
}
