package events

import (
	"context"
	"github.com/IBM/sarama"
	events2 "webook/article/events"
	"webook/internal/repository"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type HistoryReadEventConsumer struct {
	client sarama.Client
	repo   repository.HistoryRecordRepository
	l      logger.LoggerV1
}

func NewHistoryReadEventConsumer(client sarama.Client, repo repository.HistoryRecordRepository, l logger.LoggerV1) *HistoryReadEventConsumer {
	return &HistoryReadEventConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (c *HistoryReadEventConsumer) Start() error {
	// 在这里上报 prometheus 就可以
	cg, err := sarama.NewConsumerGroupFromClient("history_record", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"read_article"}, saramax.NewHandler[events2.ReadEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (c *HistoryReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t events2.ReadEvent) error {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//return c.repo.AddRecord(context.Background(), "article", t.Aid)
	return nil
}
