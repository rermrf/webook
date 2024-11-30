package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/pkg/logger"
	"webook/pkg/saramax"
	"webook/search/service"
)

type SyncDataEvent struct {
	IndexName string
	DocId     string
	Data      string
}

type SyncDataEventConsumer struct {
	svc    service.SyncService
	client sarama.Client
	l      logger.LoggerV1
}

func (s *SyncDataEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("search_sync_data",
		s.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicSyncArticle},
			saramax.NewHandler[SyncDataEvent](s.l, s.Consume))
		if err != nil {
			s.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (s *SyncDataEventConsumer) Consume(sg *sarama.ConsumerMessage,
	evt SyncDataEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.svc.InputAny(ctx, evt.IndexName, evt.DocId, evt.Data)
}
