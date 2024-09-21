package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Upsert(ctx context.Context, pArt PublishArticle) error
}

type GormArticleDao struct {
	db *gorm.DB
}

func NewGormArticleDao(db *gorm.DB) ArticleDao {
	return &GormArticleDao{
		db: db,
	}
}

// Upsert INSERT or UPDATE
func (dao GormArticleDao) Upsert(ctx context.Context, pArt PublishArticle) error {
	now := time.Now().UnixMilli()
	pArt.Ctime = now
	pArt.Utime = now
	// 插入
	// OnConflict 的意思是数据冲突了
	err := dao.db.Clauses(clause.OnConflict{
		// SQL 2003 标准
		// - 哪些列冲突
		//Columns: []clause.Column{clause.Column{Name: "id"}},
		// - 意思是数据冲突，啥也不干
		//DoNothing: true,
		// - 数据冲突了，并且符合 WHERE 条件的就会执行更新 DoUpdates
		//Where:

		//Mysql 只需要关心这个
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   pArt.Title,
			"content": pArt.Content,
			"utime":   now,
		}),
	}).Create(&pArt).Error
	// Mysql 最终的语句 INSERT xxx ON DUPLICATE KEY UPDATE xxx

	return err
}

func (dao GormArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	// 先操作制作库(表)，后操作线上库(表)
	// 在事务内部，这里采用了闭包形态
	// GORM 帮助我们管理了事物的生命周期
	// Begin，Rollback 和 Commit 都不需要我们操心
	var (
		id = art.Id
	)
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		txDao := NewGormArticleDao(tx)
		if id > 0 {
			err = txDao.UpdateById(ctx, art)
		} else {
			id, err = txDao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}

		// 新增到线上表
		return txDao.Upsert(ctx, PublishArticle{
			Article: art,
		})
	})

	return id, err
}

func (dao GormArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return 1, err
}

func (dao GormArticleDao) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	// 依赖 gorm 忽略零值的特性
	//err := dao.db.WithContext(ctx).Updates(&art).Error
	res := dao.db.WithContext(ctx).Model(&art).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"utime":   art.Utime,
	})
	// 检查是否真的更新
	if res.Error != nil {
		return res.Error
	}
	// 更新行数
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败，可能是创作者非法 id：%d，author_id：%d", art.Id, art.AuthorId)
	}
	return res.Error
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
