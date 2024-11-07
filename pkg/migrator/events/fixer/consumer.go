package fixer

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
	"webook/pkg/migrator/fixer"
	"webook/pkg/saramax"
)

type Consumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.LoggerV1
	srcFirst *fixer.Fixer[T]
	dstFirst *fixer.Fixer[T]
	topic    string
}

func NewConsumer[T migrator.Entity](
	client sarama.Client,
	l logger.LoggerV1,
	topic string,
	src *gorm.DB,
	dst *gorm.DB) (*Consumer[T], error) {
	srcFirst, err := fixer.NewFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	return &Consumer[T]{
		client:   client,
		l:        l,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
		topic:    topic,
	}, nil
}

func (c *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{c.topic}, saramax.NewHandler[events.InconsistentEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (c *Consumer[T]) Consume(msg *sarama.ConsumerMessage, t events.InconsistentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	switch t.Direction {
	case "SRC":
		return c.srcFirst.Fix(ctx, t)
	case "DST":
		return c.dstFirst.Fix(ctx, t)
	}
	return errors.New("未知的校验方向")
}
