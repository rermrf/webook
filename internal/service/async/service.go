package async

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/internal/service/sms"
	logger2 "webook/pkg/logger"
)

type Service struct {
	svc sms.Service
	// 转异步，存储发送短信的请求
	repo repository.AsyncSMSRepository
	l    logger2.LoggerV1
}

func NewService(svc sms.Service, repo repository.AsyncSMSRepository, l logger2.LoggerV1) *Service {
	res := &Service{
		svc:  svc,
		repo: repo,
		l:    l,
	}
	go func() {
		res.StartAsyncCycle()
	}()
	return res
}

// StartAsyncCycle 异步发送消息
// 这里我们没有设计退出机制，是因为没啥必要
// 因为程序停止的时候，它自然就停止了
// 原理：这是最简单的抢占式调度
func (s *Service) StartAsyncCycle() {
	for {
		s.AsyncSend()
	}
}

func (s *Service) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 抢占一个异步发送的消息
	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch err {
	case nil:
		// 执行发送
		// 也可以做成可配置的
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = s.svc.Send(ctx, as.Biz, as.Args, as.Numbers...)
		if err != nil {
			// 啥也不干
			s.l.Error("执行异步发送短信失败",
				logger2.Error(err),
				logger2.Int64("id", as.Id))
		}
		res := err == nil
		// 通知 repository 我这一次的执行结果
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			s.l.Error("执行异步发送短信成功，但是标记数据库失败",
				logger2.Error(err),
				logger2.Bool("res", res),
				logger2.Int64("id", as.Id))
		}
	case repository.ErrWaitingSMSNotFound:
		// 睡一秒
		time.Sleep(time.Second)
	default:
		// 正常来说应该是数据库那边出了问题，
		// 但是为了尽量运行，还是要继续的
		// 可以稍微睡眠，也可以不睡眠
		// 睡眠的话可以帮你规避掉短时间的网络抖动问题
		s.l.Error("抢占异步发送短信任务失败",
			logger2.Error(err))
		time.Sleep(time.Second)
	}

}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	if s.needAsync() {
		// 需要异步发送，直接转存到数据库
		err := s.repo.Add(ctx, domain.AsyncSMS{
			Biz:     biz,
			Args:    args,
			Numbers: numbers,
			// 设置可以重试三次
			RetryMax: 3,
		})
		return err
	}
	return s.svc.Send(ctx, biz, args, numbers...)
}

func (s *Service) needAsync() bool {
	// 1. 基于响应时间的，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒钟增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率
	return true
}
