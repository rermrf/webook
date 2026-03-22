package service

import (
	"context"

	"webook/im/domain"
	"webook/im/repository"
)

type ConversationService interface {
	ListConversations(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.Conversation, bool, error)
	GetConversation(ctx context.Context, conversationId string) (domain.Conversation, error)
	GetUnreadCount(ctx context.Context, userId int64) (int64, map[string]int64, error)
	IsOnline(ctx context.Context, userId int64) (bool, error)
}

type conversationService struct {
	convRepo repository.ConversationRepository
}

func NewConversationService(convRepo repository.ConversationRepository) ConversationService {
	return &conversationService{
		convRepo: convRepo,
	}
}

func (s *conversationService) ListConversations(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.Conversation, bool, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	conversations, err := s.convRepo.FindByUserId(ctx, userId, cursor, limit)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(conversations) == limit
	return conversations, hasMore, nil
}

func (s *conversationService) GetConversation(ctx context.Context, conversationId string) (domain.Conversation, error) {
	return s.convRepo.FindByConversationId(ctx, conversationId)
}

func (s *conversationService) GetUnreadCount(ctx context.Context, userId int64) (int64, map[string]int64, error) {
	byConversation, err := s.convRepo.GetUnreadCount(ctx, userId)
	if err != nil {
		return 0, nil, err
	}
	var total int64
	for _, count := range byConversation {
		total += count
	}
	return total, byConversation, nil
}

func (s *conversationService) IsOnline(ctx context.Context, userId int64) (bool, error) {
	return s.convRepo.IsOnline(ctx, userId)
}
