package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"webook/internal/service/sms"
)

// FailoverSMSService 轮询
type FailoverSMSService struct {
	// 多个服务商
	svcs []sms.Service
	idx  uint64
}

func NewFailoverSMSService(svcs []sms.Service) sms.Service {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

func (f FailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// 缺点，每次都从头开始，绝大多数请求会在 svcs[0] 就成功。负载不均衡
	// 如果 svcs 有几十个，轮询都很慢
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tpl, args, numbers...)
		if err == nil {
			// 发送成功
			return nil
		}
		// 输出日志
		// 做好监控
		log.Println(err)
	}
	return errors.New("全部都失败了")
}

func (f FailoverSMSService) SendV1(ctx context.Context, tpl string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[i%length]
		err := svc.Send(ctx, tpl, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			return err
		default:
			log.Println(err)
		}
	}
	return errors.New("全部都失败了")
}
