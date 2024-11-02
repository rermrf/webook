package dao

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
	Upsert(ctx context.Context, pArt PublishedArticle) error
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
	GetByAuthor(ctx context.Context, author int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	ListPub(ctx context.Context, startTime time.Time, offset int, limit int) ([]Article, error)
}

type GormArticleDao struct {
	db *gorm.DB
}

func NewGormArticleDao(db *gorm.DB) ArticleDao {
	return &GormArticleDao{
		db: db,
	}
}

func (dao GormArticleDao) ListPub(ctx context.Context, startTime time.Time, offset int, limit int) ([]Article, error) {
	var res []Article
	err := dao.db.WithContext(ctx).
		Where("utime < ?", startTime.UnixMilli()).
		Order("utime desc").
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (dao GormArticleDao) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var res PublishedArticle
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&res).Error
	return res, err
}

func (dao GormArticleDao) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).
		Where("id = ?", id).First(&art).Error
	return art, err
}

func (dao GormArticleDao) GetByAuthor(ctx context.Context, author int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	// SELECT * FROM XXX WHERE XX ORDER BY utime DESC
	// 在设计 order by 语句的时候，要注意让 order by 中的数据命中索引
	// 索引天然有序，因为使用了 B+ 树

	// SQL 优化的案例：早期的时候，
	// 我们的 order by 没有命中索引的，内存排序非常慢
	// 工作就是优化了这个查询，加进去了索引
	// author_id => author_id, utime 的联合索引
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", author).
		Offset(offset).
		Limit(limit).
		// 升序排序：utime ASC
		// 混合排序：
		// ctime ASC, utime DESC
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (dao GormArticleDao) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", id, author).
			Updates(map[string]any{
				"status": status,
				"Utime":  now,
			})
		if res.Error != nil {
			// 数据库有问题
			return res.Error
		}
		if res.RowsAffected != 1 {
			// 要么 ID 是错的，要么作者不对
			// 在后者情况下，就要小心，可能有人在搞你的系统
			// 用 prometheus 打点，只要频繁出现，就告警，然后手工介入排查
			return fmt.Errorf("操作非自己的文章, uid: %d, aid: %d", id, id)
		}
		return tx.Model(&PublishedArticle{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"status": status,
				"Utime":  now,
			}).Error
	})
}

// Upsert INSERT or UPDATE
func (dao GormArticleDao) Upsert(ctx context.Context, pArt PublishedArticle) error {
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
			"status":  pArt.Status,
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
		art.Id = id
		// 新增到线上表
		return txDao.Upsert(ctx, PublishedArticle(art))
	})

	return id, err
}

func (dao GormArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao GormArticleDao) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	// 依赖 gorm 忽略零值的特性
	//err := dao.db.WithContext(ctx).Updates(&art).Error
	res := dao.db.WithContext(ctx).Model(&art).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
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
	Id      int64  `gorm:"primaryKey;autoIncrement" bson:"id,omitempty"`
	Title   string `gorm:"type:varchar(1024)" bson:"title,omitempty"`
	Content string `gorm:"type:BLOB" bson:"content,omitempty"`
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
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`

	// 有些人考虑到，经常用状态来查询
	// WHERE status = xxx AND
	// 在 status 上和别的列混在一起，创建一个联合索引
	// 要看别的列究竟是什么列
	Status uint8 `bson:"status,omitempty"`
	Ctime  int64 `bson:"ctime,omitempty"`
	Utime  int64 `bson:"utime,omitempty"`
}

//func (u *Article) BeforeCreate(tx *gorm.DB) error {
//	startTime := time.Now()
//	tx.Set("start_time", startTime)
//	slog.Default().Info("这是 create 的钩子函数")
//	return nil
//}
//
//func (u *Article) AfterCreate(tx *gorm.DB) error {
//	val, _ := tx.Get("start_time")
//	startTime, ok := val.(time.Time)
//	if !ok {
//		return nil
//	}
//	duration := time.Since(startTime)
//	slog.Default().Info("这是 create 的钩子函数")
//	return nil
//}

type PublishedArticle Article
