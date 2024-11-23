package events

import (
	"context"
	"github.com/IBM/sarama"
	"strings"
	"time"
	"webook/pkg/logger"
	"webook/pkg/saramax"
	"webook/reward/domain"
	"webook/reward/service"
)

type PaymentEvent struct {
	BizTradeNO string
	Status     uint8
}

func (p PaymentEvent) ToDomainStatus() domain.RewardStatus {
	// PaymentStatusInit
	// PaymentStatusSuccess
	// PaymentStatusFailed
	// PaymentStatusRefund
	switch p.Status {
	case 1:
		return domain.RewardStatusInit
	case 2:
		return domain.RewardStatusPayed
	case 3, 4:
		return domain.RewardStatusFailed
	default:
		return domain.RewardStatusUnknown
	}
}

type PaymentEventConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	svc    service.RewardService
}

// Start 这边自己启动 goroutine
func (r *PaymentEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("reward", r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"payment_events"},
			saramax.NewHandler[PaymentEvent](r.l, r.Consume),
		)
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *PaymentEventConsumer) Consume(msg *sarama.ConsumerMessage, evt PaymentEvent) error {
	// 不是我们的，我只处理reward
	// biz_trade_no 是以 reward 开头
	if !strings.HasPrefix(string(msg.Topic), "reward") {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return r.svc.UpdateReward(ctx, evt.BizTradeNO, evt.ToDomainStatus())
}
