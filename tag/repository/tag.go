package repository

import (
	"context"
	"time"
	"webook/pkg/logger"
	"webook/tag/domain"
	"webook/tag/repository/cache"
	"webook/tag/repository/dao"
)

type CachedTagRepository struct {
	dao   dao.TagDao
	cache cache.TagCache
	l     logger.LoggerV1
}

func NewCachedTagRepository(dao dao.TagDao, cache cache.TagCache, l logger.LoggerV1) TagRepository {
	return &CachedTagRepository{dao: dao, cache: cache, l: l}
}

func (r *CachedTagRepository) CreateTag(ctx context.Context, tag domain.Tag) (int64, error) {
	id, err := r.dao.CreateTag(ctx, r.toEntity(tag))
	if err != nil {
		return 0, err
	}
	err = r.cache.Append(ctx, tag.Uid, tag)
	if err != nil {
		r.l.Error("缓存tag失败", logger.Int64("id", id), logger.Error(err))
	}
	return id, nil
}

func (r *CachedTagRepository) BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tagIds []int64) error {
	tagBizs := make([]dao.TagBiz, 0, len(tagIds))
	for _, tagId := range tagIds {
		tagBizs = append(tagBizs, dao.TagBiz{
			Tid:   tagId,
			BizId: bizId,
			Biz:   biz,
			Uid:   uid,
		})
	}
	return r.dao.CreateTagBiz(ctx, tagBizs)
}

func (r *CachedTagRepository) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	res, err := r.cache.GetTags(ctx, uid)
	if err == nil {
		return res, nil
	}
	tags, err := r.dao.GetTagsByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	res = make([]domain.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, r.toDomain(tag))
	}
	err = r.cache.Append(ctx, uid, res...)
	if err != nil {
		r.l.Error("缓存tag失败", logger.Error(err))
	}
	return res, nil
}

func (r *CachedTagRepository) GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error) {
	tags, err := r.dao.GetTagsById(ctx, ids)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, r.toDomain(tag))
	}
	return res, nil
}

func (r *CachedTagRepository) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	// 这里要不要缓存
	tags, err := r.dao.GetTagsByBiz(ctx, uid, biz, bizId)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, r.toDomain(tag))
	}
	return res, nil
}

// PreloadUserTags 在 toB 的场景下，你可以提前预加载缓存
func (repo *CachedTagRepository) PreloadUserTags(ctx context.Context) error {
	// 我们要存的是 uid => 我的所有标签 [{tag1}, {}]
	// 你这边分批次预加载
	// 数据里面取出来，调用append
	offset := 0
	batch := 100
	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		tags, err := repo.dao.GetTags(dbCtx, offset, batch)
		cancel()
		if err != nil {
			// 你也可以 continue
			return err
		}
		for _, tag := range tags {
			rctx, cancel := context.WithTimeout(ctx, time.Second)
			err = repo.cache.Append(rctx, tag.Uid, repo.toDomain(tag))
			cancel()
			if err != nil {
				continue
			}
		}
		if len(tags) < batch {
			return nil
		}
		offset = offset + batch
	}
}

func (r *CachedTagRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}

func (r *CachedTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
