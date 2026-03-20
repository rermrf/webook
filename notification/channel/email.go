package channel

import (
	"context"
	"fmt"
	"webook/notification/domain"
)

// EmailSender 邮件渠道发送器（占位实现）
type EmailSender struct{}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

func (s *EmailSender) Send(ctx context.Context, notification domain.Notification) error {
	return fmt.Errorf("email 渠道暂未实现")
}

func (s *EmailSender) BatchSend(ctx context.Context, notifications []domain.Notification) error {
	return fmt.Errorf("email 渠道暂未实现")
}
