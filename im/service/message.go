package service

import (
	"context"
	"fmt"
	"time"

	"webook/im/domain"
	"webook/im/repository"
)

type MessageService interface {
	SendMessage(ctx context.Context, msg domain.Message) (domain.Message, error)
	ListMessages(ctx context.Context, conversationId string, cursor int64, limit int) ([]domain.Message, bool, error)
	RecallMessage(ctx context.Context, userId int64, messageId string) error
	MarkAsRead(ctx context.Context, userId int64, conversationId string) error
}

type messageService struct {
	msgRepo  repository.MessageRepository
	convRepo repository.ConversationRepository
}

func NewMessageService(msgRepo repository.MessageRepository, convRepo repository.ConversationRepository) MessageService {
	return &messageService{
		msgRepo:  msgRepo,
		convRepo: convRepo,
	}
}

func (s *messageService) SendMessage(ctx context.Context, msg domain.Message) (domain.Message, error) {
	// 1. 生成会话ID
	convId := domain.GenConversationID(msg.SenderId, msg.ReceiverId)
	msg.ConversationID = convId
	msg.Status = domain.MsgStatusSent
	msg.Ctime = time.Now().UnixMilli()

	// 2. 确保会话存在
	err := s.convRepo.CreateIfNotExist(ctx, domain.Conversation{
		ConversationID: convId,
		Members:        []int64{msg.SenderId, msg.ReceiverId},
	})
	if err != nil {
		return domain.Message{}, fmt.Errorf("创建会话失败: %w", err)
	}

	// 3. 持久化消息
	id, err := s.msgRepo.Create(ctx, msg)
	if err != nil {
		return domain.Message{}, fmt.Errorf("保存消息失败: %w", err)
	}
	msg.Id = id

	// 4. 更新会话最后一条消息
	err = s.convRepo.UpdateLastMsg(ctx, convId, domain.LastMessage{
		Content:  msg.Content,
		MsgType:  msg.MsgType,
		SenderId: msg.SenderId,
		Ctime:    msg.Ctime,
	})
	if err != nil {
		return domain.Message{}, fmt.Errorf("更新最后消息失败: %w", err)
	}

	// 5. 递增接收方未读计数
	err = s.convRepo.IncrUnread(ctx, msg.ReceiverId, convId)
	if err != nil {
		return domain.Message{}, fmt.Errorf("递增未读计数失败: %w", err)
	}

	// 6. 更新双方会话排序分数
	err = s.convRepo.UpdateConvScore(ctx, msg.SenderId, convId, msg.Ctime)
	if err != nil {
		return domain.Message{}, fmt.Errorf("更新发送方会话分数失败: %w", err)
	}
	err = s.convRepo.UpdateConvScore(ctx, msg.ReceiverId, convId, msg.Ctime)
	if err != nil {
		return domain.Message{}, fmt.Errorf("更新接收方会话分数失败: %w", err)
	}

	return msg, nil
}

func (s *messageService) ListMessages(ctx context.Context, conversationId string, cursor int64, limit int) ([]domain.Message, bool, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	messages, err := s.msgRepo.FindByConversation(ctx, conversationId, cursor, limit)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(messages) == limit
	return messages, hasMore, nil
}

func (s *messageService) RecallMessage(ctx context.Context, userId int64, messageId string) error {
	msg, err := s.msgRepo.FindById(ctx, messageId)
	if err != nil {
		return fmt.Errorf("查找消息失败: %w", err)
	}
	if msg.SenderId != userId {
		return fmt.Errorf("只有发送者才能撤回消息")
	}
	return s.msgRepo.UpdateStatus(ctx, messageId, domain.MsgStatusRecalled)
}

func (s *messageService) MarkAsRead(ctx context.Context, userId int64, conversationId string) error {
	err := s.msgRepo.MarkConversationRead(ctx, conversationId, userId)
	if err != nil {
		return fmt.Errorf("标记已读失败: %w", err)
	}
	return s.convRepo.ClearUnread(ctx, userId, conversationId)
}
