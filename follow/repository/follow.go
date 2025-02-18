package repository

import (
	"context"
	"webook/follow/domain"
	"webook/follow/repository/cache"
	"webook/follow/repository/dao"
	"webook/pkg/logger"
)

type CachedFollowRepository struct {
	dao   dao.FollowDao
	cache cache.FollowCache
	l     logger.LoggerV1
}

func NewCachedFollowRepository(dao dao.FollowDao, cache cache.FollowCache, l logger.LoggerV1) FollowRepository {
	return &CachedFollowRepository{dao: dao, cache: cache, l: l}
}

func (c *CachedFollowRepository) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	// 你可以在这里缓存关注着列表的第一页
	followerList, err := c.dao.FollowRelationList(ctx, follower, offset, limit)
	if err != nil {
		return nil, err
	}
	return c.getFollowRelationList(followerList), nil
}

func (c *CachedFollowRepository) GetFollower(ctx context.Context, followee, offset, limit int64) ([]domain.FollowRelation, error) {
	followerList, err := c.dao.FollowerRelationList(ctx, followee, offset, limit)
	if err != nil {
		return nil, err
	}
	return c.getFollowRelationList(followerList), nil
}

func (c *CachedFollowRepository) FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error) {
	// follow:123:234 => 标签信息，分组信息
	fr, err := c.dao.FollowRelationDetail(ctx, follower, followee)
	if err != nil {
		return domain.FollowRelation{}, err
	}
	return c.toDomain(fr), nil
}

func (c *CachedFollowRepository) AddFollowRelation(ctx context.Context, f domain.FollowRelation) error {
	err := c.dao.CreateFollowRelation(ctx, c.toEntity(f))
	if err != nil {
		return err
	}
	// 更新缓存里面的关注了多少人，以及有多少粉丝的计数，+1
	return c.cache.Follow(ctx, f.Follower, f.Followee)
}

func (c *CachedFollowRepository) InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error {
	err := c.dao.UpdateStatus(ctx, follower, followee, dao.FollowRelationStatusInactive)
	if err != nil {
		return err
	}
	// 缓存 -1
	return c.cache.CancelFollow(ctx, follower, followee)
}

// GetFollowStatics 获取个人关注了多少人，已经粉丝数量
func (c *CachedFollowRepository) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	res, err := c.cache.StaticsInfo(ctx, uid)
	if err == nil {
		return res, nil
	}
	// 没有则去数据库获取
	res.Followers, err = c.dao.CntFollower(ctx, uid)
	if err != nil {
		return domain.FollowStatics{}, err
	}
	res.Followees, err = c.dao.CntFollowee(ctx, uid)
	if err != nil {
		return domain.FollowStatics{}, err
	}
	err = c.cache.SetStaticsInfo(ctx, uid, res)
	if err != nil {
		c.l.Error("记录关注信息错误", logger.Error(err))
	}
	return res, nil
}

func (c *CachedFollowRepository) getFollowRelationList(list []dao.FollowRelation) []domain.FollowRelation {
	res := make([]domain.FollowRelation, 0, len(list))
	for _, v := range list {
		res = append(res, c.toDomain(v))
	}
	return res
}

func (c *CachedFollowRepository) Cache() cache.FollowCache {
	return c.cache
}

func (c *CachedFollowRepository) toDomain(v dao.FollowRelation) domain.FollowRelation {
	return domain.FollowRelation{
		Followee: v.Followee,
		Follower: v.Follower,
	}
}

func (c *CachedFollowRepository) toEntity(v domain.FollowRelation) dao.FollowRelation {
	return dao.FollowRelation{
		Followee: v.Followee,
		Follower: v.Follower,
	}
}
