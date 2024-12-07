package events

import (
	"context"
	"github.com/IBM/sarama"
	"go.uber.org/atomic"
	"time"
	"webook/pkg/canalx"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/validator"
)

type MysqlBinlogConsumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.LoggerV1
	table    string
	srcToDst *validator.CanalIncrValidator[T]
	dtsToSrc *validator.CanalIncrValidator[T]
	dstFirst *atomic.Bool
}

func NewMysqlBinlogConsumer[T migrator.Entity](
	client sarama.Client,
	l logger.LoggerV1,
	table string,
	srcToDst *validator.CanalIncrValidator[T],
	dtsToSrc *validator.CanalIncrValidator[T],
	dstFirst *atomic.Bool) *MysqlBinlogConsumer[T] {
	return &MysqlBinlogConsumer[T]{
		client:   client,
		l:        l,
		table:    table,
		srcToDst: srcToDst,
		dtsToSrc: dtsToSrc,
		dstFirst: dstFirst,
	}
}

func (m *MysqlBinlogConsumer[T]) Start() error {
	//TODO implement me
	panic("implement me")
}

func (m *MysqlBinlogConsumer[T]) Consume(msg *sarama.ConsumerMessage, val canalx.Message[T]) error {
	// 是不是源表为准
	dstFirst := m.dstFirst.Load()
	var v *validator.CanalIncrValidator[T]
	// db:
	//   src:
	//     dsn: "root:root@tcp(localhost:13306)/webook"
	//   dst:
	//     dsn: "root:root@tcp(localhost:13306)/webook_intr"
	if dstFirst && val.Database == "webook_intr" {
		// 目标表为准
		// 校验，用 dst 来校验
		v = m.dtsToSrc
	} else if !dstFirst && val.Database == "webook" {
		// 源表为准，而且消息恰好来自源表
		// 校验，用 Src 校验
		v = m.srcToDst
	}
	if v != nil {
		for _, data := range val.Data {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := v.Validate(ctx, data.ID())
			cancel()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
