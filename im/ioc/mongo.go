package ioc

import (
	"context"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"webook/im/repository/dao"
)

func InitMongo() *mongo.Database {
	type Config struct {
		URI    string `yaml:"uri"`
		DBName string `yaml:"dbName"`
	}
	var cfg Config
	err := viper.UnmarshalKey("mongo", &cfg)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		panic(err)
	}
	db := client.Database(cfg.DBName)
	err = dao.InitCollections(db)
	if err != nil {
		panic(err)
	}
	return db
}
