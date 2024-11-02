package logger

import (
	"context"
	"go.uber.org/zap"
	"webook/sms/service"
)

type Service struct {
	svc service.Service
}

// Send 装饰器 Aop 打印日志
func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	zap.L().Debug("发送短信",
		zap.String("biz", biz),
		zap.Strings("args", args))
	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		zap.L().Error("发送短信出现异常", zap.Error(err))
	}
	zap.L().Error("短信发送成功")
	return err
}
