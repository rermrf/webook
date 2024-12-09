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

// FeedEvent 异步调用，数据同步接口
type FeedEvent struct {
	// Type 是我内部定义，我发给不同业务方
	Type string
	// 业务方具体的数据
	// 点赞需要的key：
	// liker
	// liked
	// biz + bizId
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
