package connpool

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/atomic"
	"gorm.io/gorm"
)

type Mysql2Mongo struct {
	db      *gorm.DB
	mdb     *mongo.Database
	pattern atomic.String
}
