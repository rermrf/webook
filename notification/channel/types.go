package channel

import (
	"context"
	"webook/notification/domain"
)

// Sender 渠道发送器接口
type Sender interface {
	// Send 发送单条通知
	Send(ctx context.Context, notification domain.Notification) error
	// BatchSend 批量发送通知
	BatchSend(ctx context.Context, notifications []domain.Notification) error
}
