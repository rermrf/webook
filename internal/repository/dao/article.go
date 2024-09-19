package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
}

type GormArticleDao struct {
	db *gorm.DB
}

func NewGormArticleDao(db *gorm.DB) ArticleDao {
	return &GormArticleDao{
		db: db,
	}
}

func (dao GormArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.Create(&art).Error
	return 1, err
}

// Article 制作库
type Article struct {
	Id      int64  `gorm:"primaryKey;autoIncrement"`
	Title   string `gorm:"type:varchar(1024)"`
	Content string `gorm:"type:BLOB"`
	// 如何设置索引
	// 在帖子这里，什么样查询场景
	// 1. 对于创作者来说，看草稿箱，看到所有自己的文章
	// SELECT * FROM articles WHERE author_id = 123 ORDER BY `ctime` DESC;
	// 产品经理告诉你，按照帖子的创建时间倒序排序
	// 2. 单独查询某一篇
	// SELECT * FROM articles WHERE id = 1
	// - 在 author_id 和 ctime 上创建联合索引
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`

	// TODO: 学习 Explain 命令

	// - 在 author_id 上创建索引
	AuthorId int64 `gorm:"index"`
	Ctime    int64
	Utime    int64
}
