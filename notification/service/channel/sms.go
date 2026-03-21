package channel

import (
	"context"
	"webook/notification/domain"
	"webook/notification/repository"
)

// SMSSender 短信渠道发送器
// 使用 SMSProvider 发送短信，通过 TemplateRepository 查找短信服务商模板 ID。
type SMSSender struct {
	provider SMSProvider
	tplRepo  repository.TemplateRepository
}

func NewSMSSender(provider SMSProvider, tplRepo repository.TemplateRepository) *SMSSender {
	return &SMSSender{
		provider: provider,
		tplRepo:  tplRepo,
	}
}

func (s *SMSSender) Send(ctx context.Context, notification domain.Notification) error {
	// 查找 SMS 渠道对应的模板
	tpl, err := s.tplRepo.FindByTemplateIdAndChannel(ctx, notification.TemplateId, domain.ChannelSMS)
	if err != nil {
		return err
	}
	// 使用服务商模板 ID 和通知参数发送短信
	return s.provider.Send(ctx, tpl.SMSProviderTemplateId, notification.TemplateParams, notification.Receiver)
}

func (s *SMSSender) BatchSend(ctx context.Context, notifications []domain.Notification) error {
	for _, n := range notifications {
		if err := s.Send(ctx, n); err != nil {
			return err
		}
	}
	return nil
}
