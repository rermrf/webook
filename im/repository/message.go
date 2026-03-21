package repository

import (
	"context"

	"webook/im/domain"
	"webook/im/repository/dao"
)

type MessageRepository interface {
	Create(ctx context.Context, msg domain.Message) (string, error)
	FindByConversation(ctx context.Context, conversationId string, cursor int64, limit int) ([]domain.Message, error)
	FindById(ctx context.Context, id string) (domain.Message, error)
	UpdateStatus(ctx context.Context, id string, status domain.MsgStatus) error
	MarkConversationRead(ctx context.Context, conversationId string, receiverId int64) error
}

type messageRepository struct {
	dao dao.MessageDAO
}

func NewMessageRepository(dao dao.MessageDAO) MessageRepository {
	return &messageRepository{dao: dao}
}

func (r *messageRepository) Create(ctx context.Context, msg domain.Message) (string, error) {
	return r.dao.Insert(ctx, r.toEntity(msg))
}

func (r *messageRepository) FindByConversation(ctx context.Context, conversationId string, cursor int64, limit int) ([]domain.Message, error) {
	msgs, err := r.dao.FindByConversation(ctx, conversationId, cursor, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Message, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, r.toDomain(m))
	}
	return result, nil
}

func (r *messageRepository) FindById(ctx context.Context, id string) (domain.Message, error) {
	msg, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.Message{}, err
	}
	return r.toDomain(msg), nil
}

func (r *messageRepository) UpdateStatus(ctx context.Context, id string, status domain.MsgStatus) error {
	return r.dao.UpdateStatus(ctx, id, uint8(status))
}

func (r *messageRepository) MarkConversationRead(ctx context.Context, conversationId string, receiverId int64) error {
	err := r.dao.UpdateStatusBatch(ctx, conversationId, receiverId, uint8(domain.MsgStatusSent), uint8(domain.MsgStatusRead))
	if err != nil {
		return err
	}
	return r.dao.UpdateStatusBatch(ctx, conversationId, receiverId, uint8(domain.MsgStatusDelivered), uint8(domain.MsgStatusRead))
}

func (r *messageRepository) toEntity(msg domain.Message) dao.Message {
	return dao.Message{
		ConversationID: msg.ConversationID,
		SenderId:       msg.SenderId,
		ReceiverId:     msg.ReceiverId,
		MsgType:        uint8(msg.MsgType),
		Content:        msg.Content,
		Status:         uint8(msg.Status),
		Ctime:          msg.Ctime,
	}
}

func (r *messageRepository) toDomain(msg dao.Message) domain.Message {
	return domain.Message{
		Id:             msg.ID.Hex(),
		ConversationID: msg.ConversationID,
		SenderId:       msg.SenderId,
		ReceiverId:     msg.ReceiverId,
		MsgType:        domain.MsgType(msg.MsgType),
		Content:        msg.Content,
		Status:         domain.MsgStatus(msg.Status),
		Ctime:          msg.Ctime,
	}
}
