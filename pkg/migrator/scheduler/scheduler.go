package scheduler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"sync"
	"time"
	"webook/pkg/ginx"
	"webook/pkg/gormx/connpool"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
	"webook/pkg/migrator/validator"
)

// Scheduler 用来统一管理这个迁移过程
// 他不是必须的，可以理解为这是为了方便用户操作而引入的
type Scheduler[T migrator.Entity] struct {
	lock sync.Mutex
	src  *gorm.DB
	dst  *gorm.DB
	// 业务使用的 DoubleWritePool
	pool       *connpool.DoubleWritePool
	l          logger.LoggerV1
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer

	// 如果要允许多个全量校验同时并行
	fulls map[string]func()
}

func NewScheduler[T migrator.Entity](
	src *gorm.DB,
	dst *gorm.DB,
	pool *connpool.DoubleWritePool,
	l logger.LoggerV1,
	producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{
		src:     src,
		dst:     dst,
		pool:    pool,
		l:       l,
		pattern: connpool.PatternSrcOnly,
		cancelFull: func() {

		},
		cancelIncr: func() {

		},
		producer: producer,
	}
}

// RegisterRoutes 也不是必须的，可以考虑利用配置中心，监听配置中心的变化
// 也可以把全量校验，增量校验做成分布式任务，利用分布式任务调度平台来调度
func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	// 将这个暴露成为 HTTP接口
	server.POST("/src_only", ginx.Wrap(s.l, s.SrcOnly))
	server.POST("/src_first", ginx.Wrap(s.l, s.SrcFirst))
	server.POST("/dst_first", ginx.Wrap(s.l, s.DstFirst))
	server.POST("/dst_only", ginx.Wrap(s.l, s.DstOnly))
	server.POST("/full/start", ginx.Wrap(s.l, s.StartFullValidation))
	server.POST("/full/stop", ginx.Wrap(s.l, s.StopFullValidation))
	server.POST("/incr/stop", ginx.Wrap(s.l, s.StopIncrementValidation))
	server.POST("/incr/start", ginx.WrapBody[StartIncrementValidationRequest](s.l, s.StartIncrementValidation))
}

// SrcOnly 只读源表
func (s *Scheduler[T]) SrcOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.UpdatePattern(connpool.PatternSrcOnly)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) SrcFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	s.cancelFull()
	s.pattern = connpool.PatternSrcFirst
	s.pool.UpdatePattern(connpool.PatternSrcFirst)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) DstOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.UpdatePattern(connpool.PatternDstOnly)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) DstFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.UpdatePattern(connpool.PatternDstFirst)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) StartFullValidation(c *gin.Context) (ginx.Result, error) {
	// 可以考虑去重问题
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{}, nil
	}

	go func() {
		var ctx context.Context
		ctx, s.cancelFull = context.WithCancel(context.Background())

		// 先取消上一次的
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("退出全量校验", logger.Error(err))
		}
	}()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) StopFullValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) StartIncrementValidation(c *gin.Context, req StartIncrementValidationRequest) (ginx.Result, error) {
	// 开启增量校验
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelIncr
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		}, nil
	}
	v.Incr().Utime(req.Utime).SleepInterval(time.Duration(req.Interval) * time.Millisecond)
	var ctx context.Context
	ctx, s.cancelIncr = context.WithCancel(context.Background())

	go func() {
		// 一样先取消上一次的
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("退出增量校验", logger.Error(err))
		}
	}()
	return ginx.Result{
		Msg: "启动增量校验成功",
	}, nil
}

type StartIncrementValidationRequest struct {
	Utime int64 `json:"utime"`
	// 毫秒数
	// json 不能正确处理 time.Duration 类型
	Interval int64 `json:"interval"`
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, s.l, s.producer, "SRC"), nil
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		return validator.NewValidator[T](s.dst, s.src, s.l, s.producer, "DST"), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}
