package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	logger2 "webook/pkg/logger"
)

type Handler[T any] struct {
	l  logger2.LoggerV1
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](l logger2.LoggerV1, fn func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	return &Handler[T]{
		l:  l,
		fn: fn,
	}
}

func (h Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			// 打日志
			h.l.Error("反序列化消息失败",
				logger2.Error(err),
				logger2.String("topic", msg.Topic),
				logger2.Int32("partition", msg.Partition),
				logger2.Int64("offset", msg.Offset),
			)
			continue
		}
		//err = h.fn(msg, t)

		// 在这里执行重试
		for i := 0; i < 3; i++ {
			err = h.fn(msg, t)
			if err == nil {
				break
			}
			h.l.Error("处理消息失败",
				logger2.Error(err),
				logger2.String("topic", msg.Topic),
				logger2.Int32("partition", msg.Partition),
				logger2.Int64("offset", msg.Offset),
			)
		}

		if err != nil {
			h.l.Error("处理消息失败-重试次数上限",
				logger2.Error(err),
				logger2.String("topic", msg.Topic),
				logger2.Int32("partition", msg.Partition),
				logger2.Int64("offset", msg.Offset),
			)
		} else {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}
