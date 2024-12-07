package validator

import (
	"context"
	"gorm.io/gorm"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

// CanalIncrValidator 借助 canal 来执行增量校验的方法
type CanalIncrValidator[T migrator.Entity] struct {
	baseValidator
}

func NewCanalIncrValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, l logger.LoggerV1, p events.Producer, direction string) *CanalIncrValidator[T] {
	return &CanalIncrValidator[T]{
		baseValidator: baseValidator{
			base:      base,
			target:    target,
			l:         l,
			p:         p,
			direction: direction,
		},
	}
}

// Validate 一次校验一条
func (v *CanalIncrValidator[T]) Validate(ctx context.Context, id int64) error {
	var base T
	err := v.base.WithContext(ctx).Where("id = ?", id).First(&base).Error
	switch err {
	case nil:
		// 找到了，在找到 target 进行对比
		var target T
		err1 := v.target.WithContext(ctx).Where("id = ?", id).First(&target).Error
		switch err1 {
		case nil:
			// target 也找着了
			if !base.Equal(T(target)) {
				// 如果不相等，再产生一条消息，去修复数据
				v.notify(ctx, id, events.InconsistentEventTypeNEQ)
			}
		case gorm.ErrRecordNotFound:
			// target 中找不到数据
			v.notify(ctx, id, events.InconsistentEventTypeTargetMissing)
		default:
			return err
		}
	case gorm.ErrRecordNotFound:
		// 收到消息的时候，这条数据已经没有了
		// 一样去找
		var target T
		err1 := v.target.WithContext(ctx).Where("id = ?", id).First(&target).Error
		switch err1 {
		case nil:
			// 源表没了，目标表找到了
			v.notify(ctx, id, events.InconsistentEventTypeBaseMissing)
		case gorm.ErrRecordNotFound:
			// 两边都没了，什么都不需要干
		default:
			return err
		}
	default:
		return err
	}

	return nil
}
