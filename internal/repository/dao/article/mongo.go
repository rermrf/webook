package article

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{
				"author_id": 1,
				"ctime":     1,
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().CreateMany(ctx, index)
	return err
}

type MongoArticleDao struct {
	database *mongo.Database   // 代表 webook
	col      *mongo.Collection // 代表制作库
	liveCol  *mongo.Collection // 代表线上库
	node     *snowflake.Node
}

func NewMongoArticleDao(database *mongo.Database, node *snowflake.Node) ArticleDao {
	return &MongoArticleDao{
		database: database,
		col:      database.Collection("articles"),
		liveCol:  database.Collection("published_articles"),
		node:     node,
	}
}

func (m *MongoArticleDao) ListPub(ctx context.Context, startTime time.Time, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoArticleDao) GetByAuthor(ctx context.Context, author int64, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoArticleDao) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoArticleDao) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	id := m.node.Generate().Int64()
	art.Id = id
	art.Utime = now
	art.Ctime = now
	_, err := m.col.InsertOne(ctx, art)
	return id, err
}

func (m *MongoArticleDao) UpdateById(ctx context.Context, article Article) error {
	// 操作制作库
	filter := bson.M{"id": article.Id, "author_id": article.AuthorId}
	update := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title":   article.Title,
		"content": article.Content,
		"status":  article.Status,
		"utime":   time.Now().UnixMilli(),
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (m *MongoArticleDao) Sync(ctx context.Context, article Article) (int64, error) {
	// 没法引入事务的概念
	// 保存制作库
	var (
		id  = article.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, article)
	} else {
		id, err = m.Insert(ctx, article)
	}
	if err != nil {
		return 0, err
	}
	article.Id = id
	// 保存到线上库, upsert
	now := time.Now().UnixMilli()
	//update := bson.E{"$set", article}
	//upsert := bson.E{"$setOnInsert", bson.D{bson.E{"ctime", now}}}
	article.Utime = now
	upsertV1 := bson.M{
		// 更新，如果不存在，就是插入
		"$set": PublishedArticle(article),
		"$setOnInsert": bson.M{
			// 在插入的时候要插入 ctime
			"ctime": now,
		},
	}
	filter := bson.M{"id": article.Id}
	_, err = m.liveCol.UpdateOne(ctx, filter,
		//bson.D{update, upsert},
		upsertV1,
		options.Update().SetUpsert(true))
	return id, err
}

func (m *MongoArticleDao) Upsert(ctx context.Context, pArt PublishedArticle) error {
	//TODO implement me
	panic("implement me")
}

func (m *MongoArticleDao) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}

//func ToUpdate(vals map[string]interface{}) bson.M {
//	return vals
//}
//
//func ToFilter(vals map[string]interface{}) bson.D {
//	var res bson.D
//	for k, v := range vals {
//		res = append(res, bson.E{Key: k, Value: v})
//	}
//	return res
//}
//
//func Set(vals map[string]interface{}) bson.M {
//	return bson.M{"$set": bson.M{vals}}
//}
