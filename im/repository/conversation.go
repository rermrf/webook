package repository

import (
	"context"
	"time"

	"webook/im/domain"
	"webook/im/repository/cache"
	"webook/im/repository/dao"
)

type ConversationRepository interface {
	CreateIfNotExist(ctx context.Context, conv domain.Conversation) error
	FindByConversationId(ctx context.Context, conversationId string) (domain.Conversation, error)
	FindByUserId(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.Conversation, error)
	UpdateLastMsg(ctx context.Context, conversationId string, lastMsg domain.LastMessage) error
	GetUnreadCount(ctx context.Context, userId int64) (map[string]int64, error)
	IncrUnread(ctx context.Context, userId int64, conversationId string) error
	ClearUnread(ctx context.Context, userId int64, conversationId string) error
	UpdateConvScore(ctx context.Context, userId int64, conversationId string, score int64) error
}

type conversationRepository struct {
	convDAO dao.ConversationDAO
	cache   cache.IMCache
}

func NewConversationRepository(convDAO dao.ConversationDAO, cache cache.IMCache) ConversationRepository {
	return &conversationRepository{
		convDAO: convDAO,
		cache:   cache,
	}
}

func (r *conversationRepository) CreateIfNotExist(ctx context.Context, conv domain.Conversation) error {
	now := time.Now().UnixMilli()
	return r.convDAO.Upsert(ctx, dao.Conversation{
		ConversationID: conv.ConversationID,
		Members:        conv.Members,
		Ctime:          now,
		Utime:          now,
	})
}

func (r *conversationRepository) FindByConversationId(ctx context.Context, conversationId string) (domain.Conversation, error) {
	conv, err := r.convDAO.FindByConversationId(ctx, conversationId)
	if err != nil {
		return domain.Conversation{}, err
	}
	return r.toDomain(conv), nil
}

func (r *conversationRepository) FindByUserId(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.Conversation, error) {
	convs, err := r.convDAO.FindByUserId(ctx, userId, cursor, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Conversation, 0, len(convs))
	for _, c := range convs {
		result = append(result, r.toDomain(c))
	}
	return result, nil
}

func (r *conversationRepository) UpdateLastMsg(ctx context.Context, conversationId string, lastMsg domain.LastMessage) error {
	now := time.Now().UnixMilli()
	return r.convDAO.UpdateLastMsg(ctx, conversationId, dao.LastMessage{
		Content:  lastMsg.Content,
		MsgType:  uint8(lastMsg.MsgType),
		SenderId: lastMsg.SenderId,
		Ctime:    lastMsg.Ctime,
	}, now)
}

func (r *conversationRepository) GetUnreadCount(ctx context.Context, userId int64) (map[string]int64, error) {
	return r.cache.GetUnread(ctx, userId)
}

func (r *conversationRepository) IncrUnread(ctx context.Context, userId int64, conversationId string) error {
	return r.cache.IncrUnread(ctx, userId, conversationId)
}

func (r *conversationRepository) ClearUnread(ctx context.Context, userId int64, conversationId string) error {
	return r.cache.ClearUnread(ctx, userId, conversationId)
}

func (r *conversationRepository) UpdateConvScore(ctx context.Context, userId int64, conversationId string, score int64) error {
	return r.cache.UpdateConvScore(ctx, userId, conversationId, score)
}

func (r *conversationRepository) toDomain(conv dao.Conversation) domain.Conversation {
	return domain.Conversation{
		Id:             conv.ID.Hex(),
		ConversationID: conv.ConversationID,
		Members:        conv.Members,
		LastMsg: domain.LastMessage{
			Content:  conv.LastMsg.Content,
			MsgType:  domain.MsgType(conv.LastMsg.MsgType),
			SenderId: conv.LastMsg.SenderId,
			Ctime:    conv.LastMsg.Ctime,
		},
		Ctime: conv.Ctime,
		Utime: conv.Utime,
	}
}
