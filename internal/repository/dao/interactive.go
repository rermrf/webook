package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

type InteractiveDao interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error)
	GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error)
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
	//GetItems() ([]ColletctionItem, error)
}

type GORMInteractiveDao struct {
	db *gorm.DB
}

func NewGORMInteractiveDao(db *gorm.DB) InteractiveDao {
	return &GORMInteractiveDao{
		db: db,
	}
}

func (dao *GORMInteractiveDao) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	// 很多新手都会犯的错误：先查询数据库中的阅读数，再阅读数加一更新到数据库

	// 有一个没考虑到，就是，我可能根本就没这一行
	// 事实上这里就是一个 upsert 的语义
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		// Mysql 不写
		//Columns:
		DoUpdates: clause.Assignments(map[string]any{
			"read_cnt": gorm.Expr("read_cnt + ?", 1),
			"utime":    time.Now().UnixMilli(),
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

func (dao *GORMInteractiveDao) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	// 同时记录点赞以及更新点赞计数
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先准备插入点赞记录
		// 有可能已经点赞过了
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"utime":  now,
				"Status": 1,
			}),
		}).Create(&UserLikeBiz{
			Biz:    biz,
			BizId:  bizId,
			Uid:    uid,
			Status: 1,
			Ctime:  now,
			Utime:  now,
		}).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(ctx).Clauses(clause.OnConflict{
			// Mysql 不写
			//Columns:
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("like_cnt + ?", 1),
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   bizId,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
		return err
	})
}

func (dao *GORMInteractiveDao) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	// WithContext(ctx) 控制事物超时
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 软删除点赞记录
		// 2. 减点赞数量
		err := tx.Model(&UserLikeBiz{}).Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).Updates(map[string]any{
			"utime":  now,
			"Status": 0,
		}).Error
		if err != nil {
			return err
		}
		// 这边命中了索引，然后没找到，所以不会加锁
		return tx.Model(&Interactive{}).Where("biz = ? AND biz_id = ?", biz, bizId).Updates(map[string]any{
			"utime":    now,
			"like_cnt": gorm.Expr("like_cnt - ?", 1),
		}).Error
	})
}

func (dao *GORMInteractiveDao) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Utime = now
	cb.Ctime = now
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 插入收藏项
		err := dao.db.WithContext(ctx).Create(&cb).Error
		if err != nil {
			return err
		}
		// 更新数量
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"collect_cnt": gorm.Expr("collect_cnt + ?", 1),
				"utime":       now,
			}),
		}).Create(&Interactive{
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
			Biz:        cb.Biz,
			BizId:      cb.BizId,
		}).Error
	})
}

func (dao *GORMInteractiveDao) GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND uid = ? AND status = ?", biz, bizId, uid, 1).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDao) GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND uid = ? AND status = ?", biz, bizId, uid, 1).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDao) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ?", biz, bizId).First(&res).Error
	return res, err
}

// Interactive 假如我要查找点赞数量前 100 的：
// SELECT * FROM (SELECCT biz, biz_id, COUNT(*) as cnt FROM `interactives` GROUP BY biz, biz_id) ORDER BY cnt limit 100;
// 实时查找，性能贼差，上面的语句，就是全表扫描
// 高性能
// 面试的标准答案：zset
// 但是，面试标准不够有特色，烂大街
// 可以考虑别的方案：
// 1. 定时计算
// 1.1 定时计算 + 本地缓存
// 2. 优化版的 zset，定时筛选 + 实时 zset 计算
type Interactive struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`
	// 业务标识符
	// 同一个资源，在这里应该只有一行
	// 也就是说在 bizId 和 biz 上创建联合唯一索引
	// 1. bizId, biz 优先选择这个，因为 bizId 的区分度更高
	// 2. biz, bizId。如果有 WHERE biz = xx 这种查询条件（不带 bizId）的，就只能这种
	//
	// 总结：联合索引的列的顺序：查询条件，区分度
	BizId int64  `gorm:"uniqueIndex:biz_id_type"`
	Biz   string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"`
	// 阅读计数
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

// InteractiveV1 对写更友好
// Interactive 对读更加友好
//type InteractiveV1 struct {
//	I'd    int64 `gorm:"primaryKey;autoIncrement"`
//	BizId int64
//	Biz   string
//	// 阅读计数
//	Cnt     int64
//	CntType string
//	Ctime   int64
//	Utime   int64
//}

// UserLikeBiz 用户点赞的某个东西
type UserLikeBiz struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	// WHERE uid = ? AND biz_id = ? AND biz = ?
	// 来判定你有没有点赞
	// 这里的联合索引顺序：
	// 1. 如果用户要看看自己点赞过那些，uid 在前
	// WHERE uid = ?
	// 2. 如果我的点赞数量，需要通过这里来比较/纠正, biz_id 和 biz 在前
	// SELECT count(*) WHERE biz = ? and biz_id = ?
	Biz   string `gorm:"uniqueIndex:uid_biz_id_type;type:varchar(128)"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_id_type"`

	Uid   int64 `gorm:"uniqueIndex:uid_biz_id_type"`
	Ctime int64
	Utime int64

	// 软删除
	// 这个状态是存储状态，纯粹用于软删除的，业务层面上是没有感知的
	// 0-代表删除，1-代表有效
	Status uint8
}

// UserCollectionBiz 收藏的东西
type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`
	// 收藏夹 ID
	// 作为关联关系中的外键，需要索引
	Cid int64 `gorm:"index"`
	// 搜藏的东西的id
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	// 这算是一个冗余，因为正常来说维持着 Uid 就可以
	Uid   int64 `gorm:"uniqueIndex:biz_type_id_uid"`
	Ctime int64
	Utime int64
}

// 假如我有一个需求，查询到收藏夹的信息，和收藏夹里面的资源
// SELECT c.id as cid, c.name as cname, uc.biz_id as biz_id, uc.biz as biz
// FROM `collection` as c JOIN `user_collection_biz` as uc
// ON c.id = uc.cid
// WHERE c.id IN (1,2,3)

//type ColletctionItem struct {
//	Cid   int64
//	Cname string
//	BizId int64
//	Biz   string
//}
//
//func (dao *GORMInteractiveDao) GetItems() ([]ColletctionItem, error) {
//	var items []ColletctionItem
//	err := dao.db.Raw("", 1, 2, 3).Find(&items).Error
//	return items, err
//}
