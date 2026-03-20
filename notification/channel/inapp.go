package channel

import (
	"context"
	"webook/notification/domain"
	"webook/notification/repository"
)

// InAppSender 站内通知发送器
// 通过 NotificationRepository 创建通知记录，
// 仓储层内部已处理缓存更新和 SSE 推送。
type InAppSender struct {
	repo repository.NotificationRepository
}

func NewInAppSender(repo repository.NotificationRepository) *InAppSender {
	return &InAppSender{
		repo: repo,
	}
}

func (s *InAppSender) Send(ctx context.Context, notification domain.Notification) error {
	_, err := s.repo.Create(ctx, notification)
	return err
}

func (s *InAppSender) BatchSend(ctx context.Context, notifications []domain.Notification) error {
	_, err := s.repo.BatchCreate(ctx, notifications)
	return err
}
