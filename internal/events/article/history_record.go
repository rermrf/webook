package article

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/internal/pkg/logger"
	"webook/internal/pkg/saramax"
	"webook/internal/repository"
)

type HistoryReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewHistoryReadEventConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.LoggerV1) *HistoryReadEventConsumer {
	return &HistoryReadEventConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (c *HistoryReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"read_article"}, saramax.NewHandler[ReadEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (c *HistoryReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return c.repo.AddRecord(ctx, "article", t.Aid)
}
