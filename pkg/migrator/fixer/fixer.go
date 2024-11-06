package fixer

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

type Fixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

// Fix 直接覆盖
// TODO 改成批量
func (f *Fixer[T]) Fix(ctx context.Context, evt events.InconsistentEvent) error {
	var t T
	err := f.base.WithContext(ctx).Where("id = ?", evt.Id).First(&t).Error
	switch err {
	case nil:
		// base 有数据
		// 修复数据的时候，可以考虑 WHERE base.utime > target.utime
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	case gorm.ErrRecordNotFound:
		// base 没了
		return f.target.WithContext(ctx).Where("id = ?", evt.Id).Delete(&t).Error
	default:
		return err
	}
}

func (f *Fixer[T]) FixV2(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeTargetMissing:
		// 目标表插入缺失数据
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.Id).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// base 也删除了这个记录
			return nil
		case nil:
			return f.target.Create(&t).Error
		default:
			return err
		}

	case events.InconsistentEventTypeNEQ:
		// 不相等，更新目标表数据
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.Id).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// base 删除了这个记录，所以 target也删除
			return f.target.WithContext(ctx).Where("id = ? ", evt.Id).Delete(&t).Error
		case nil:
			return f.target.WithContext(ctx).Updates(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		// 删除目标表中多余的数据
		return f.target.WithContext(ctx).Where("id = ? ", evt.Id).Delete(new(T)).Error
	default:
		return errors.New("未知的不一致类型")
	}
}

// base 和 target 在校验时候的数据，到你修复的时候就变了
func (f *Fixer[T]) FixV1(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeTargetMissing, events.InconsistentEventTypeNEQ:
		// 目标表插入缺失数据
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.Id).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// base 也删除了这个记录
			return f.target.WithContext(ctx).Where("id = ? ", evt.Id).Delete(new(T)).Error
		case nil:
			return f.target.Clauses(clause.OnConflict{
				// 要更新全部的列
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		// 删除目标表中多余的数据
		return f.target.WithContext(ctx).Where("id = ? ", evt.Id).Delete(new(T)).Error
	default:
		return errors.New("未知的不一致类型")
	}
}
