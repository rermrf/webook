package dao

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Conversation struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	ConversationID string             `bson:"conversation_id"`
	Members        []int64            `bson:"members"`
	LastMsg        LastMessage         `bson:"last_msg"`
	Ctime          int64              `bson:"ctime"`
	Utime          int64              `bson:"utime"`
}

type LastMessage struct {
	Content  string `bson:"content"`
	MsgType  uint8  `bson:"msg_type"`
	SenderId int64  `bson:"sender_id"`
	Ctime    int64  `bson:"ctime"`
}

type ConversationDAO interface {
	Upsert(ctx context.Context, conv Conversation) error
	FindByConversationId(ctx context.Context, conversationId string) (Conversation, error)
	FindByUserId(ctx context.Context, userId int64, cursor int64, limit int) ([]Conversation, error)
	UpdateLastMsg(ctx context.Context, conversationId string, lastMsg LastMessage, utime int64) error
}

type conversationDAO struct {
	col *mongo.Collection
}

func NewConversationDAO(db *mongo.Database) ConversationDAO {
	return &conversationDAO{
		col: db.Collection("conversations"),
	}
}

func (d *conversationDAO) Upsert(ctx context.Context, conv Conversation) error {
	filter := bson.M{"conversation_id": conv.ConversationID}
	update := bson.M{
		"$setOnInsert": bson.M{
			"ctime":   conv.Ctime,
			"members": conv.Members,
		},
		"$set": bson.M{
			"utime": conv.Utime,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := d.col.UpdateOne(ctx, filter, update, opts)
	return err
}

func (d *conversationDAO) FindByConversationId(ctx context.Context, conversationId string) (Conversation, error) {
	var conv Conversation
	err := d.col.FindOne(ctx, bson.M{"conversation_id": conversationId}).Decode(&conv)
	return conv, err
}

func (d *conversationDAO) FindByUserId(ctx context.Context, userId int64, cursor int64, limit int) ([]Conversation, error) {
	filter := bson.M{"members": userId}
	if cursor > 0 {
		filter["utime"] = bson.M{"$lt": cursor}
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "utime", Value: -1}}).
		SetLimit(int64(limit))

	cur, err := d.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var convs []Conversation
	if err = cur.All(ctx, &convs); err != nil {
		return nil, err
	}
	return convs, nil
}

func (d *conversationDAO) UpdateLastMsg(ctx context.Context, conversationId string, lastMsg LastMessage, utime int64) error {
	filter := bson.M{"conversation_id": conversationId}
	update := bson.M{
		"$set": bson.M{
			"last_msg": lastMsg,
			"utime":    utime,
		},
	}
	_, err := d.col.UpdateOne(ctx, filter, update)
	return err
}
