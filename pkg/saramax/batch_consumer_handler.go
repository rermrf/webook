package saramax

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"time"
	logger2 "webook/pkg/logger"
)

// BatchHandler 批量消费接口
type BatchHandler[T any] struct {
	l  logger2.LoggerV1
	fn func(msgs []*sarama.ConsumerMessage, t []T) error
	// 用 option 模式来设置
	batchSize     int
	batchDuration time.Duration
}

func NewBatchHandler[T any](l logger2.LoggerV1, fn func(msgs []*sarama.ConsumerMessage, t []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{
		l:             l,
		fn:            fn,
		batchSize:     10,
		batchDuration: time.Second,
	}
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 批量消费
	msgsCh := claim.Messages()
	batchSize := b.batchSize
	ctx, cancel := context.WithTimeout(context.Background(), b.batchDuration)
	for {
		done := false
		msgs := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 超时
				done = true
			case msg, ok := <-msgsCh:
				if !ok {
					cancel()
					// 消费者被关闭
					return nil
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列化失败",
						logger2.Error(err),
						logger2.String("topic", msg.Topic),
						logger2.Int32("partition", msg.Partition),
						logger2.Int64("offset", msg.Offset))
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}
		cancel()
		if len(msgs) == 0 {
			continue
		}
		err := b.fn(msgs, ts)
		if err != nil {
			b.l.Error("调用业务批量接口失败",
				logger2.Error(err))

			// 还要继续消费
		}
		for _, msg := range msgs {
			session.MarkMessage(msg, "")
		}
	}
}
