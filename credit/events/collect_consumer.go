package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/credit/service"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type CollectEventConsumer struct {
	client sarama.Client
	svc    service.CreditService
	l      logger.LoggerV1
}

func NewCollectEventConsumer(client sarama.Client, svc service.CreditService, l logger.LoggerV1) *CollectEventConsumer {
	return &CollectEventConsumer{
		client: client,
		svc:    svc,
		l:      l,
	}
}

func (c *CollectEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("credit-collect", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"collect_events"}, saramax.NewHandler[CollectEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (c *CollectEventConsumer) Consume(msg *sarama.ConsumerMessage, evt CollectEvent) error {
	// 只处理收藏操作，取消收藏不处理
	if evt.Action != "collect" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, _, _, err := c.svc.EarnCredit(ctx, evt.Uid, "collect", evt.BizId)
	if err != nil {
		c.l.Error("处理收藏积分事件失败",
			logger.Int64("uid", evt.Uid),
			logger.Int64("bizId", evt.BizId),
			logger.Error(err))
		return err
	}
	return nil
}
