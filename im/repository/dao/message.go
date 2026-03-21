package dao

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	ConversationID string             `bson:"conversation_id"`
	SenderId       int64              `bson:"sender_id"`
	ReceiverId     int64              `bson:"receiver_id"`
	MsgType        uint8              `bson:"msg_type"`
	Content        string             `bson:"content"`
	Status         uint8              `bson:"status"`
	Ctime          int64              `bson:"ctime"`
}

type MessageDAO interface {
	Insert(ctx context.Context, msg Message) (string, error)
	FindByConversation(ctx context.Context, conversationId string, cursor int64, limit int) ([]Message, error)
	FindById(ctx context.Context, id string) (Message, error)
	UpdateStatus(ctx context.Context, id string, status uint8) error
	UpdateStatusBatch(ctx context.Context, conversationId string, receiverId int64, fromStatus, toStatus uint8) error
}

type messageDAO struct {
	col *mongo.Collection
}

func NewMessageDAO(db *mongo.Database) MessageDAO {
	return &messageDAO{
		col: db.Collection("messages"),
	}
}

func (d *messageDAO) Insert(ctx context.Context, msg Message) (string, error) {
	result, err := d.col.InsertOne(ctx, msg)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (d *messageDAO) FindByConversation(ctx context.Context, conversationId string, cursor int64, limit int) ([]Message, error) {
	filter := bson.M{"conversation_id": conversationId}
	if cursor > 0 {
		filter["ctime"] = bson.M{"$lt": cursor}
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "ctime", Value: -1}}).
		SetLimit(int64(limit))

	cur, err := d.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var msgs []Message
	if err = cur.All(ctx, &msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}

func (d *messageDAO) FindById(ctx context.Context, id string) (Message, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Message{}, err
	}
	var msg Message
	err = d.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&msg)
	return msg, err
}

func (d *messageDAO) UpdateStatus(ctx context.Context, id string, status uint8) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = d.col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"status": status}})
	return err
}

func (d *messageDAO) UpdateStatusBatch(ctx context.Context, conversationId string, receiverId int64, fromStatus, toStatus uint8) error {
	filter := bson.M{
		"conversation_id": conversationId,
		"receiver_id":     receiverId,
		"status":          fromStatus,
	}
	update := bson.M{"$set": bson.M{"status": toStatus}}
	_, err := d.col.UpdateMany(ctx, filter, update)
	return err
}
