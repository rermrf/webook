package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/interactive/repository"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type InteractiveReadBatchConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewInteractiveReadBatchConsumer(client sarama.Client, l logger.LoggerV1, repo repository.InteractiveRepository) *InteractiveReadBatchConsumer {
	return &InteractiveReadBatchConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (k *InteractiveReadBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", k.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"read_article"}, saramax.NewBatchHandler[ReadEvent](k.l, k.Consume))
		if err != nil {
			k.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (k *InteractiveReadBatchConsumer) Consume(msgs []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ids := make([]int64, 0, len(ts))
	bizs := make([]string, 0, len(ts))
	for _, evt := range ts {
		ids = append(ids, evt.Aid)
		bizs = append(bizs, "article")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := k.repo.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		k.l.Error("批量增加阅读计数失败",
			logger.Field{Key: "ids", Value: ids},
			logger.Error(err))
	}
	return nil
}
