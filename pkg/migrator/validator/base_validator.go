package validator

import (
	"context"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator/events"
)

type baseValidator struct {
	// 基准
	base *gorm.DB
	// 目标
	target *gorm.DB
	l      logger.LoggerV1

	p         events.Producer
	direction string
}

func (v *baseValidator) notify(ctx context.Context, id int64, typ string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	err := v.p.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		Id:        id,
		Direction: v.direction,
		Type:      typ,
	})
	cancel()
	if err != nil {
		// 发现数据不一致并且发送失败
		// 可以重试，但是重试也会失败，记日志，告警，手动修复
		// 可以直接忽略，下一轮修复和校验又会找出来
		v.l.Error("发送数据不一致的消息失败", logger.Error(err))
	}
}
