package validator

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

// 最好就是全量校验用id，增量校验用canal

// Validator T 必须实现了 Entity 接口
type Validator[T migrator.Entity] struct {
	// 基准
	base *gorm.DB
	// 目标
	target *gorm.DB
	l      logger.LoggerV1

	p         events.Producer
	direction string
	batchSize int

	//highLoad *atomic.Value

	utime int64
	// <= 0 说明直接退出校验循环
	// > 0 真的 sleep
	sleepInterval time.Duration

	fromBase func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, l logger.LoggerV1, p events.Producer, direction string) *Validator[T] {
	//highLoad := &atomic.Value{}
	//highLoad.Store(false)
	//go func() {
	//	// 在这里，去查询数据库的状态
	//	// 校验代码不太可能是性能瓶颈，真正的性能瓶颈一般在数据库
	//	// 也可以结合本地的CPU和内存判定
	//}()
	res := &Validator[T]{
		base:      base,
		target:    target,
		l:         l,
		p:         p,
		direction: direction,
		batchSize: 100,
		//highLoad:  highLoad,
	}
	res.fromBase = res.fullFromBase
	return res
}

func (v *Validator[T]) SleepInterval(i time.Duration) *Validator[T] {
	v.sleepInterval = i
	return v
}

func (v *Validator[T]) Utime(utime int64) *Validator[T] {
	v.utime = utime
	return v
}

// Incr 增量模式
func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

// Validate 调用者可以通过 ctx 来控制程序退出
// 全量检验，是不是一条条比对？
// 所以要从数据库里面查询出来
// utime 上面至少要有一个索引，并且 utime 必须是第一列
func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := 0
	for {
		//if v.highLoad.Load().(bool) {
		//	// 挂起
		//}

		// 找到了 base 中的数据
		// 例如 .Order("id DESC")，每次插入数据，就会导致你的 offset 不准了
		// 假如我的表没有 id 这个列怎么办？
		// 找个类似的列，比如说 ctime
		// TODO 改成批量
		src, err := v.fromBase(ctx, offset)
		switch err {
		case context.DeadlineExceeded, context.Canceled:
			// 超时或者被人取消
			return nil
		case nil:
			// 查到了数据
			// 要去 target 里面找到对应的数据
			var dst T
			err := v.target.WithContext(ctx).Where("id = ?", src.ID()).First(&dst).Error
			switch err {
			case context.DeadlineExceeded, context.Canceled:
				// 超时或者被人取消
				return nil
			case nil:
				// 找到了，开始比较
				// 怎么比较？
				// 1. 利用反射进行比较
				// 这个原则上可以
				//if reflect.DeepEqual(src, dst) {
				//
				//}
				if !src.Equal(dst) {
					// 不相等
					// 这时候，上报给 Kafka，告知数据不一致
					v.notify(ctx, src.ID(), events.InconsistentEventTypeNEQ)
				}

			case gorm.ErrRecordNotFound:
				// target 少了数据
				v.notify(ctx, src.ID(), events.InconsistentEventTypeTargetMissing)
			default:
				// 要不要汇报，数据不一致？
				// 两种做法
				// 1. 我认为，大概率数据是一致的，我记录一下日志，下一条
				v.l.Error("查询 target 数据失败", logger.Error(err))
				// 2. 我认为，处于保险起见，我应该报数据不一致，试着去修一下
				// 如果真的不一致，没事，修它
				// 如果假的不一致（也就是数据一致），也无伤大雅
			}

		case gorm.ErrRecordNotFound:
			// 比完了，没数据了，全量校验结束了
			// 同时支持全量校验和增量校验，这里就不能直接返回
			// 在这里，需要考虑：有些情况下，用户希望退出，有些情况下。用户希望继续
			// 当用户希望继续的时候，就要 sleep 一下
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		default:
			v.l.Error("校验数据，查询 base 出错", logger.Error(err))
		}
		offset++
	}
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	// 找到了 base 中的数据
	// 例如 .Order("id DESC")，每次插入数据，就会导致你的 offset 不准了
	// 假如我的表没有 id 这个列怎么办？
	// 找个类似的列，比如说 ctime
	// TODO 改成批量
	err := v.base.WithContext(dbCtx).
		Offset(offset).
		Order("id").
		First(&src).Error
	return src, err
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	// 找到了 base 中的数据
	// 例如 .Order("id DESC")，每次插入数据，就会导致你的 offset 不准了
	// 假如我的表没有 id 这个列怎么办？
	// 找个类似的列，比如说 ctime
	// TODO 改成批量
	err := v.base.WithContext(dbCtx).
		Where("utime > ?", v.utime).
		Offset(offset).
		Order("utime").
		First(&src).Error
	return src, err
}

// 摆脱对 T 的依赖
//func (v *Validator[T]) incrFromBaseV1(ctx context.Context, offset int) (T, error) {
//	rows, err := v.base.WithContext(ctx).Where("utime > ?", v.utime).Offset(offset).Order("utime").Rows()
//	cols, err := rows.Columns()
//	// 所有列的值
//	vals := make([]interface{}, len(cols))
//	vals.Scan(vals...)
//	return vals
//}

// 理论上来说，可以利用 count 加速这个过程
// 举个例子，假如你初始化目标表的数据是 昨天的 23.59.59 导出来的
// 那么你可以 COUNT(*) WHERE ctime < 今天的零点，count 如果相等，说明没删除
// 这一步大多数效果很好，尤其是那些软删除的
// 如果 count 不一致，那么接下来，理论上来说，还可以分段 count
// 比如说，我先 count 第一个月的数据，一旦有数据删除了，你还的一条一条来

// A utime = 昨天
// A 在 base 里面，今天删了，A 在 target 里面，还是昨天
// 这个地方，可以考虑不用 utime
// A 在删除之前，已经被修改过了，那么 A 在 target 里面的 utime 就是今天了
func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	// 先找 target，再找 base，找出 base 中已经被删除的
	// 理论上来说，就是 target 里面一条一条找
	offset := 0
	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		var dstTs []T
		err := v.target.WithContext(dbCtx).
			Where("utime > ?", v.utime).
			Select("id").
			Offset(offset).
			Limit(v.batchSize).
			Order("utime").
			Find(&dstTs).Error
		cancel()
		// 未找到任何数据时
		if len(dstTs) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		switch err {
		case context.DeadlineExceeded, context.Canceled:
			// 超时或者被人取消
			return nil
		// 正常来说，gorm 在 Find 方法接收的是切片的时候，不会返回 gorm.ErrRecordNotFound
		case gorm.ErrRecordNotFound:
			// 没数据了，直接返回
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		case nil:
			var ids []int64
			for _, dstT := range dstTs {
				ids = append(ids, dstT.ID())
			}
			var srcTs []T
			err = v.base.Where("id IN (?)", ids).Find(&srcTs).Error
			switch err {
			case context.DeadlineExceeded, context.Canceled:
				// 超时或者被人取消
				return nil
			case nil:
				var srcIds []int64
				for _, srcT := range srcTs {
					srcIds = append(srcIds, srcT.ID())
				}
				// 计算差集
				// 也就是 src 中没有的
				diff := slice.DiffSet(ids, srcIds)
				v.notifyBaseMissing(ctx, diff)
			case gorm.ErrRecordNotFound:
				// 全没有
				v.notifyBaseMissing(ctx, ids)
			default:
				// 记录日志
			}
		default:
			// 记录日志，continue
			v.l.Error("查询 target 失败", logger.Error(err))
		}
		offset += len(dstTs)
		if len(dstTs) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			//continue
		}
	}
}

func (v *Validator[T]) notifyBaseMissing(ctx context.Context, ids []int64) {
	for _, id := range ids {
		v.notify(ctx, id, events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notify(ctx context.Context, id int64, typ string) {
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
