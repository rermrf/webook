package job

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/semaphore"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/pkg/logger"
)

type Executor interface {
	// Name Executor 叫什么
	Name() string
	// Exec ctx 是整个任务的上下文
	// 当从 ctx.Done 有信号的时候，就需要考虑结束执行
	// 真正去执行一个任务
	Exec(ctx context.Context, j domain.Job) error
}

type HttpExecutor struct {
}

func (h *HttpExecutor) Name() string {
	return "http"
}

func (h *HttpExecutor) Exec(ctx context.Context, j domain.Job) error {
	type Config struct {
		Endpoint string
		Method   string
	}
	var cfg Config
	err := json.Unmarshal([]byte(j.Cfg), &cfg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(cfg.Method, cfg.Endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("执行失败")
	}
	return nil
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: make(map[string]func(ctx context.Context, j domain.Job) error),
	}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) RegisterFuncs(name string, f func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = f
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未知任务，你是否注册了？%s", j.Name)
	}
	return fn(ctx, j)
}

// Scheduler 调度器
type Scheduler struct {
	execs   map[string]Executor
	svc     service.JobService
	l       logger.LoggerV1
	limiter *semaphore.Weighted
}

func NewScheduler(l logger.LoggerV1, svc service.JobService) *Scheduler {
	return &Scheduler{
		l:       l,
		svc:     svc,
		execs:   map[string]Executor{},
		limiter: semaphore.NewWeighted(200),
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			// 退出调度循环
			return ctx.Err()
		}
		// 这里最多开出200个 goroutine，不够时会阻塞在这里
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		// 一次调度数据库查询时间
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 不能 return
			// 继续下一轮
			s.l.Error("抢占任务失败", logger.Error(err))
		}

		exec, ok := s.execs[j.Executor]
		if !ok {
			// DEBUG 的时候最好中断
			// 线上继续
			s.l.Error("未找到对应的执行器", logger.String("executor", j.Executor))
			continue
		}

		// 接下来继续执行
		// 怎么执行
		go func() {
			defer func() {
				s.limiter.Release(1)
				er := j.CancelFunc()
				if er != nil {
					s.l.Error("释放任务失败", logger.Error(er), logger.Int64("id", j.Id))
				}
			}()
			// 异步执行，不要阻塞主调度循环
			// 这里要考虑任务的超时控制
			er := exec.Exec(ctx, j)
			if er != nil {
				// 也可以考虑这里重试
				s.l.Error("任务执行失败", logger.Error(er))
			}
			// 执行完毕
			// 要不要考虑下一次调度？
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			er = s.svc.ResetNextTime(ctx, j)
			if er != nil {
				s.l.Error("设置下一次执行时间失败", logger.Error(er))
			}
		}()

	}
}
