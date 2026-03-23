package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/credit/service"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type LikeEventConsumer struct {
	client sarama.Client
	svc    service.CreditService
	l      logger.LoggerV1
}

func NewLikeEventConsumer(client sarama.Client, svc service.CreditService, l logger.LoggerV1) *LikeEventConsumer {
	return &LikeEventConsumer{
		client: client,
		svc:    svc,
		l:      l,
	}
}

func (c *LikeEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("credit-like", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"like_events"}, saramax.NewHandler[LikeEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (c *LikeEventConsumer) Consume(msg *sarama.ConsumerMessage, evt LikeEvent) error {
	// 只处理点赞操作，取消点赞不处理
	if evt.Action != "like" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, _, _, err := c.svc.EarnCredit(ctx, evt.Uid, "like", evt.BizId)
	if err != nil {
		c.l.Error("处理点赞积分事件失败",
			logger.Int64("uid", evt.Uid),
			logger.Int64("bizId", evt.BizId),
			logger.Error(err))
		return err
	}
	return nil
}
