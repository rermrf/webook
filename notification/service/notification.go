package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"webook/notification/service/channel"
	"webook/notification/domain"
	"webook/notification/repository"
)

// NotificationService 通知服务接口
type NotificationService interface {
	// Send 普通发送
	Send(ctx context.Context, n domain.Notification) (int64, error)
	// BatchSend 批量发送（接收已按 receiver 拆分好的通知列表）
	BatchSend(ctx context.Context, ns []domain.Notification) ([]int64, error)
	// Prepare TCC 事务 - 预提交
	Prepare(ctx context.Context, req domain.PrepareRequest) (int64, int64, error)
	// Confirm TCC 事务 - 确认
	Confirm(ctx context.Context, key string) error
	// Cancel TCC 事务 - 取消
	Cancel(ctx context.Context, key string) error
	// List 站内通知查询
	List(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error)
	// ListByGroup 按分组查询
	ListByGroup(ctx context.Context, userId int64, group domain.NotificationGroup, offset, limit int) ([]domain.Notification, error)
	// ListUnread 查询未读通知
	ListUnread(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error)
	// MarkAsRead 标记已读
	MarkAsRead(ctx context.Context, userId int64, ids []int64) error
	// MarkAllAsRead 标记全部已读
	MarkAllAsRead(ctx context.Context, userId int64) error
	// GetUnreadCount 获取未读数统计
	GetUnreadCount(ctx context.Context, userId int64) (domain.UnreadCount, error)
	// Delete 删除通知
	Delete(ctx context.Context, userId int64, id int64) error
	// DeleteAll 删除用户所有通知
	DeleteAll(ctx context.Context, userId int64) error
}

type notificationService struct {
	senders map[domain.Channel]channel.Sender
	repo    repository.NotificationRepository
	txRepo  repository.TransactionRepository
	tplSvc  TemplateService
}

func NewNotificationService(
	senders map[domain.Channel]channel.Sender,
	repo repository.NotificationRepository,
	txRepo repository.TransactionRepository,
	tplSvc TemplateService,
) NotificationService {
	return &notificationService{
		senders: senders,
		repo:    repo,
		txRepo:  txRepo,
		tplSvc:  tplSvc,
	}
}

func (s *notificationService) Send(ctx context.Context, n domain.Notification) (int64, error) {
	// 1. 生成 key（如果为空）
	if n.Key == "" {
		n.Key = uuid.New().String()
	}
	// 2. 站内通知：将 receiver 解析为 userId
	if n.Channel == domain.ChannelInApp {
		uid, err := strconv.ParseInt(n.Receiver, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("站内通知 receiver 必须为用户ID: %w", err)
		}
		n.UserId = uid
	}
	// 3. 设置初始状态
	n.Status = domain.NotificationStatusInit
	// 4. 渲染模板内容
	if n.TemplateId != "" {
		content, err := s.tplSvc.Render(ctx, n.TemplateId, n.Channel, n.TemplateParams)
		if err != nil {
			return 0, fmt.Errorf("渲染模板失败: %w", err)
		}
		n.Content = content
	}
	// 5. 路由到对应渠道的 sender
	sender, ok := s.senders[n.Channel]
	if !ok {
		return 0, fmt.Errorf("不支持的渠道: %s", n.Channel)
	}
	if err := sender.Send(ctx, n); err != nil {
		return 0, err
	}
	return 0, nil
}

func (s *notificationService) BatchSend(ctx context.Context, ns []domain.Notification) ([]int64, error) {
	if len(ns) == 0 {
		return nil, nil
	}
	if len(ns) == 1 {
		id, err := s.Send(ctx, ns[0])
		if err != nil {
			return nil, err
		}
		return []int64{id}, nil
	}
	// 批量处理：对每条通知执行 Send 流程中的准备工作
	for i := range ns {
		if ns[i].Key == "" {
			ns[i].Key = uuid.New().String()
		}
		if ns[i].Channel == domain.ChannelInApp {
			uid, err := strconv.ParseInt(ns[i].Receiver, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("站内通知 receiver 必须为用户ID: %w", err)
			}
			ns[i].UserId = uid
		}
		ns[i].Status = domain.NotificationStatusInit
		if ns[i].TemplateId != "" {
			content, err := s.tplSvc.Render(ctx, ns[i].TemplateId, ns[i].Channel, ns[i].TemplateParams)
			if err != nil {
				return nil, fmt.Errorf("渲染模板失败: %w", err)
			}
			ns[i].Content = content
		}
	}
	// 按渠道分组发送
	channelGroups := make(map[domain.Channel][]domain.Notification)
	for _, n := range ns {
		channelGroups[n.Channel] = append(channelGroups[n.Channel], n)
	}
	var allIds []int64
	for ch, group := range channelGroups {
		sender, ok := s.senders[ch]
		if !ok {
			return nil, fmt.Errorf("不支持的渠道: %s", ch)
		}
		if err := sender.BatchSend(ctx, group); err != nil {
			return nil, err
		}
	}
	return allIds, nil
}

func (s *notificationService) Prepare(ctx context.Context, req domain.PrepareRequest) (int64, int64, error) {
	n := req.Notification
	// 1. 生成 key（如果为空）
	if n.Key == "" {
		n.Key = uuid.New().String()
	}
	// 2. 站内通知：解析 receiver 为 userId
	if n.Channel == domain.ChannelInApp {
		uid, err := strconv.ParseInt(n.Receiver, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("站内通知 receiver 必须为用户ID: %w", err)
		}
		n.UserId = uid
	}
	// 3. 设置初始状态
	n.Status = domain.NotificationStatusInit
	// 4. 创建通知记录
	notificationId, err := s.repo.Create(ctx, n)
	if err != nil {
		return 0, 0, fmt.Errorf("创建通知记录失败: %w", err)
	}
	// 5. 创建事务记录
	timeout := req.CheckBackTimeoutMs
	if timeout <= 0 {
		timeout = 30000 // 默认 30 秒
	}
	tx := domain.Transaction{
		NotificationId:     notificationId,
		Key:                n.Key,
		BizId:              req.BizId,
		Status:             domain.TransactionStatusPrepared,
		CheckBackTimeoutMs: timeout,
		NextCheckTime:      time.Now().Add(time.Duration(timeout) * time.Millisecond).UnixMilli(),
		MaxRetry:           3,
	}
	txId, err := s.txRepo.Create(ctx, tx)
	if err != nil {
		return 0, 0, fmt.Errorf("创建事务记录失败: %w", err)
	}
	return notificationId, txId, nil
}

func (s *notificationService) Confirm(ctx context.Context, key string) error {
	// 1. 查找事务
	tx, err := s.txRepo.FindByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("查找事务失败: %w", err)
	}
	// 2. 检查事务状态
	if tx.Status != domain.TransactionStatusPrepared {
		return fmt.Errorf("事务状态不是已预提交，当前状态: %s", tx.Status)
	}
	// 3. 更新事务状态为已确认
	if err = s.txRepo.UpdateStatus(ctx, key, domain.TransactionStatusConfirmed); err != nil {
		return fmt.Errorf("更新事务状态失败: %w", err)
	}
	// 4. 更新通知状态为发送中
	if err = s.repo.UpdateStatus(ctx, tx.NotificationId, domain.NotificationStatusSending); err != nil {
		return fmt.Errorf("更新通知状态失败: %w", err)
	}
	// 5. 查找通知
	n, err := s.repo.FindByKeyAndChannel(ctx, key, domain.ChannelInApp)
	if err != nil {
		// 尝试其他渠道
		n, err = s.repo.FindByKeyAndChannel(ctx, key, domain.ChannelSMS)
		if err != nil {
			n, err = s.repo.FindByKeyAndChannel(ctx, key, domain.ChannelEmail)
			if err != nil {
				return fmt.Errorf("查找通知失败: %w", err)
			}
		}
	}
	// 6. 渲染模板
	if n.TemplateId != "" {
		content, er := s.tplSvc.Render(ctx, n.TemplateId, n.Channel, n.TemplateParams)
		if er != nil {
			_ = s.repo.UpdateStatus(ctx, tx.NotificationId, domain.NotificationStatusFailed)
			return fmt.Errorf("渲染模板失败: %w", er)
		}
		n.Content = content
	}
	// 7. 路由到对应渠道发送
	sender, ok := s.senders[n.Channel]
	if !ok {
		_ = s.repo.UpdateStatus(ctx, tx.NotificationId, domain.NotificationStatusFailed)
		return fmt.Errorf("不支持的渠道: %s", n.Channel)
	}
	if err = sender.Send(ctx, n); err != nil {
		_ = s.repo.UpdateStatus(ctx, tx.NotificationId, domain.NotificationStatusFailed)
		return fmt.Errorf("发送通知失败: %w", err)
	}
	// 8. 更新通知状态为已发送
	return s.repo.UpdateStatus(ctx, tx.NotificationId, domain.NotificationStatusSent)
}

func (s *notificationService) Cancel(ctx context.Context, key string) error {
	// 1. 查找事务
	tx, err := s.txRepo.FindByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("查找事务失败: %w", err)
	}
	// 2. 检查事务状态
	if tx.Status != domain.TransactionStatusPrepared {
		return fmt.Errorf("事务状态不是已预提交，当前状态: %s", tx.Status)
	}
	// 3. 更新事务状态为已取消
	if err = s.txRepo.UpdateStatus(ctx, key, domain.TransactionStatusCancelled); err != nil {
		return fmt.Errorf("更新事务状态失败: %w", err)
	}
	// 4. 更新通知状态为失败
	return s.repo.UpdateStatus(ctx, tx.NotificationId, domain.NotificationStatusFailed)
}

func (s *notificationService) List(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error) {
	return s.repo.FindByUserId(ctx, userId, offset, limit)
}

func (s *notificationService) ListByGroup(ctx context.Context, userId int64, group domain.NotificationGroup, offset, limit int) ([]domain.Notification, error) {
	return s.repo.FindByUserIdAndGroup(ctx, userId, group, offset, limit)
}

func (s *notificationService) ListUnread(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error) {
	return s.repo.FindUnreadByUserId(ctx, userId, offset, limit)
}

func (s *notificationService) MarkAsRead(ctx context.Context, userId int64, ids []int64) error {
	return s.repo.MarkAsRead(ctx, userId, ids)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userId int64) error {
	return s.repo.MarkAllAsRead(ctx, userId)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userId int64) (domain.UnreadCount, error) {
	return s.repo.GetUnreadCount(ctx, userId)
}

func (s *notificationService) Delete(ctx context.Context, userId int64, id int64) error {
	return s.repo.Delete(ctx, userId, id)
}

func (s *notificationService) DeleteAll(ctx context.Context, userId int64) error {
	return s.repo.DeleteAll(ctx, userId)
}
