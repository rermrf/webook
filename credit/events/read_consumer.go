package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/credit/service"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type ReadEventConsumer struct {
	client sarama.Client
	svc    service.CreditService
	l      logger.LoggerV1
}

func NewReadEventConsumer(client sarama.Client, svc service.CreditService, l logger.LoggerV1) *ReadEventConsumer {
	return &ReadEventConsumer{
		client: client,
		svc:    svc,
		l:      l,
	}
}

func (c *ReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("credit-read", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"read_article"}, saramax.NewHandler[ReadEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (c *ReadEventConsumer) Consume(msg *sarama.ConsumerMessage, evt ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// 使用 aid 作为 bizId，确保幂等性
	_, _, _, err := c.svc.EarnCredit(ctx, evt.Uid, "read", evt.Aid)
	if err != nil {
		c.l.Error("处理阅读积分事件失败",
			logger.Int64("uid", evt.Uid),
			logger.Int64("aid", evt.Aid),
			logger.Error(err))
		return err
	}
	return nil
}
