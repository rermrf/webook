package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/pkg/logger"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	RemoveCollectionItem(ctx context.Context, biz string, bizId int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
	//AddRecord(ctx context.Context, biz string, aid int64) error
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDao
	cache cache.InteractiveCache
	l     logger.LoggerV1
}

func NewCachedInteractiveRepository(dao dao.InteractiveDao, cache cache.InteractiveCache, l logger.LoggerV1) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

// BatchIncrReadCnt bizs 和 ids 的长度必须一致
func (c *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	// 在这里要不要检测 bizs 和 ids 的长度是否相等？
	err := c.dao.BatchIncrReadCnt(ctx, bizs, bizIds)
	if err != nil {
		return err
	}
	// 一样的批量的去修改 redis，所以要去改 lua 脚本
	//c.cache.IncrReadCntIfPresent()
	// TODO：新的lua脚本或者用pipeline
	return nil
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	// 先更新到数据库，在更新缓存
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	//go func() {
	//	c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
	//}()
	//return err
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	// 先插入点赞,然后更新点赞计数，然后更新缓存
	err := c.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	// 要不要考虑缓存收藏夹
	// 以及收藏夹里面的内容
	// 如果用户会频繁访问他的收藏夹，那么就应该缓存，不然就不需要
	// 一个东西要不要缓存，就看用户会不会频繁访问（反复访问）
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Cid:    cid,
		Biz:    biz,
		BizId:  bizId,
		Uid:    uid,
		Status: 1,
	})
	if err != nil {
		return err
	}
	// 更新收藏个数 (有多少个人收藏了这个 biz + biz_id)
	return c.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) RemoveCollectionItem(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.DeleteCollectionBiz(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrCollectCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 要从缓存里拿出阅读数，点赞数和收藏数
	intr, err := c.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, nil
	}
	// 在这里查询数据库
	daoIntr, err := c.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	intr = c.toDomain(daoIntr)
	go func() {
		er := c.cache.Set(ctx, biz, bizId, intr)
		// 记录日志
		if er != nil {
			c.l.Error("回写缓存失败",
				logger.String("biz", biz),
				logger.Int64("bizId", bizId),
				logger.Error(er))
		}
	}()
	return intr, nil
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrs, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return c.toDomains(intrs), nil
}

//func (c *CachedInteractiveRepository) GetCollection() (domain.Collection, error) {
//	items, err := c.dao.GetItems()
//	if err != nil {
//		return domain.Collection{}, err
//	}
//	// 用 items 来构建一个 Collection
//	return domain.Collection{
//		Name: items[0].Cname,
//	}, nil
//}

func (c *CachedInteractiveRepository) toDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
	}
}

func (c *CachedInteractiveRepository) toDomains(intrs []dao.Interactive) []domain.Interactive {
	result := make([]domain.Interactive, len(intrs))
	for i, intr := range intrs {
		result[i] = domain.Interactive{
			Biz:        intr.Biz,
			BizId:      intr.BizId,
			ReadCnt:    intr.ReadCnt,
			LikeCnt:    intr.LikeCnt,
			CollectCnt: intr.CollectCnt,
		}
	}
	return result
}
