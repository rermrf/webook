package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/feed/domain"
	"webook/feed/service"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

const topicFeedEvent = "feed_event"

type FeedEvent struct {
	Type     string
	Metadata map[string]string
}

type FeedEventConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	svc    service.FeedService
}

func NewFeedEventConsumer(client sarama.Client, l logger.LoggerV1, svc service.FeedService) *FeedEventConsumer {
	return &FeedEventConsumer{client: client, l: l, svc: svc}
}

func (f *FeedEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("feed_event", f.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{topicFeedEvent}, saramax.NewHandler[FeedEvent](f.l, f.Consume))
		if err != nil {
			f.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (f *FeedEventConsumer) Consume(msg *sarama.ConsumerMessage, evt FeedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return f.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: evt.Type,
		Ext:  evt.Metadata,
	})
}
