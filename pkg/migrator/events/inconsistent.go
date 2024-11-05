package events

import (
	"context"
)

type Producer interface {
	ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error
}

type InconsistentEvent struct {
	Id int64
	// 用什么来修，取值为 SRC意味着，以源表为准，取值 DST 以目标表为准
	Direction string
	// 有些时候，一些观测，或者一些第三方，需要知道，是什么引起的不一致
	// 因为他要去 DEBUG
	// 这个是可选的
	Type string
}

const (
	// InconsistentEventTypeTargetMissing 校验的目标数据，缺了这一条
	InconsistentEventTypeTargetMissing = "target_missing"
	// InconsistentEventTypeNEQ 不相等
	InconsistentEventTypeNEQ         = "neq"
	InconsistentEventTypeBaseMissing = "base_missing"
)

//type Fixer struct {
//	base   *gorm.DB
//	target *gorm.DB
//}
