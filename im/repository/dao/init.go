package dao

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitCollections(db *mongo.Database) error {
	ctx := context.Background()

	msgCol := db.Collection("messages")
	_, err := msgCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "conversation_id", Value: 1}, {Key: "ctime", Value: -1}}},
		{Keys: bson.D{{Key: "sender_id", Value: 1}, {Key: "ctime", Value: -1}}},
		{Keys: bson.D{{Key: "receiver_id", Value: 1}, {Key: "status", Value: 1}}},
	})
	if err != nil {
		return err
	}

	convCol := db.Collection("conversations")
	_, err = convCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "conversation_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "members", Value: 1}, {Key: "utime", Value: -1}}},
	})
	return err
}
