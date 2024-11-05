package validator

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"go.uber.org/atomic"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

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

	highLoad *atomic.Value
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, l logger.LoggerV1, p events.Producer, direction string) *Validator[T] {
	highLoad := &atomic.Value{}
	highLoad.Store(false)
	go func() {
		// 在这里，去查询数据库的状态
		// 校验代码不太可能是性能瓶颈，真正的性能瓶颈一般在数据库
		// 也可以结合本地的CPU和内存判定
	}()
	return &Validator[T]{
		base:      base,
		target:    target,
		l:         l,
		p:         p,
		direction: direction,
		highLoad:  highLoad,
	}
}

func (v *Validator[T]) Validate(ctx context.Context) {
	v.validateBaseToTarget(ctx)
	v.validateTargetToBase(ctx)
}

// Validate 调用者可以通过 ctx 来控制程序退出
// 全量检验，是不是一条条比对？
// 蓑衣要从数据库里面查询出来
func (v *Validator[T]) validateBaseToTarget(ctx context.Context) {
	offset := -1
	for {
		//if v.highLoad.Load().(bool) {
		//	// 挂起
		//}
		// 进来就更新 offset，比较好控制
		// 因为后面有很多的 continue 和 return
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		offset++
		var src T
		// 找到了 base 中的数据
		// 例如 .Order("id DESC")，每次插入数据，就会导致你的 offset 不准了
		// 假如我的表没有 id 这个列怎么办？
		// 找个类似的列，比如说 ctime
		err := v.base.WithContext(dbCtx).Offset(offset).Order("id").First(&src).Error
		cancel()
		switch err {
		case nil:
			// 查到了数据
			// 要去 target 里面找到对应的数据
			var dst T
			err := v.target.Where("id = ?", src.ID()).First(&dst).Error
			switch err {
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
				continue
				// 2. 我认为，处于保险起见，我应该报数据不一致，试着去修一下
				// 如果真的不一致，没事，修它
				// 如果假的不一致（也就是数据一致），也无伤大雅

			}

		case gorm.ErrRecordNotFound:
			// 比完了，没数据了，全量校验结束了
			return
		default:
			v.l.Error("校验数据，查询 base 出错", logger.Error(err))
			continue
		}
	}
}

// 理论上来说，可以利用 count 加速这个过程
// 举个例子，假如你初始化目标表的数据是 昨天的 23.59.59 导出来的
// 那么你可以 COUNT(*) WHERE ctime < 今天的零点，count 如果相等，说明没删除
// 这一步大多数效果很好，尤其是那些软删除的
// 如果 count 不一致，那么接下来，理论上来说，还可以分段 count
// 比如说，我先 count 第一个月的数据，一旦有数据删除了，你还的一条一条来
func (v *Validator[T]) validateTargetToBase(ctx context.Context) {
	// 先找 target，再找 base，找出 base 中已经被删除的
	// 理论上来说，就是 target 里面一条一条找
	offset := -v.batchSize
	for {
		offset += v.batchSize
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		var dstTs []T
		err := v.target.WithContext(dbCtx).Select("id").Offset(offset).Limit(v.batchSize).Order("id").Find(&dstTs).Error
		cancel()
		if len(dstTs) == 0 {
			return
		}
		switch err {
		case gorm.ErrRecordNotFound:
			// 没数据了，直接返回
			return
		case nil:
			var ids []int64
			for _, dstT := range dstTs {
				ids = append(ids, dstT.ID())
			}
			var srcTs []T
			err = v.base.Where("id IN (?)", ids).Find(&srcTs).Error
			switch err {
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
				continue
			}

		default:
			// 记录日志，continue
			continue
		}
		if len(dstTs) < v.batchSize {
			// 没数据了
			return
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
