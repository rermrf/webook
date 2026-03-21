package repository

import (
	"context"
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
	// 新标签创建后，清除全局标签缓存
	er := r.cache.DelAllTags(ctx)
	if er != nil {
		r.l.Error("清除全局标签缓存失败", logger.Int64("id", id), logger.Error(er))
	}
	return id, nil
}

func (r *CachedTagRepository) BindTagToBiz(ctx context.Context, biz string, bizId int64, tagIds []int64) error {
	tagBizs := make([]dao.TagBiz, 0, len(tagIds))
	for _, tagId := range tagIds {
		tagBizs = append(tagBizs, dao.TagBiz{
			Tid:   tagId,
			BizId: bizId,
			Biz:   biz,
		})
	}
	err := r.dao.CreateTagBiz(ctx, tagBizs)
	if err != nil {
		return err
	}
	// 清除该资源的标签缓存
	er := r.cache.DelBizTags(ctx, biz, bizId)
	if er != nil {
		r.l.Error("清除资源标签缓存失败", logger.Error(er))
	}
	return nil
}

func (r *CachedTagRepository) GetTags(ctx context.Context) ([]domain.Tag, error) {
	res, err := r.cache.GetAllTags(ctx)
	if err == nil {
		return res, nil
	}
	tags, err := r.dao.GetAllTags(ctx)
	if err != nil {
		return nil, err
	}
	res = make([]domain.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, r.toDomain(tag))
	}
	er := r.cache.SetAllTags(ctx, res)
	if er != nil {
		r.l.Error("缓存全局标签失败", logger.Error(er))
	}
	return res, nil
}

func (r *CachedTagRepository) GetTagById(ctx context.Context, id int64) (domain.Tag, error) {
	tag, err := r.dao.GetTagById(ctx, id)
	if err != nil {
		return domain.Tag{}, err
	}
	return r.toDomain(tag), nil
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

func (r *CachedTagRepository) GetBizTags(ctx context.Context, biz string, bizId int64) ([]domain.Tag, error) {
	res, err := r.cache.GetBizTags(ctx, biz, bizId)
	if err == nil {
		return res, nil
	}
	tags, err := r.dao.GetTagsByBiz(ctx, biz, bizId)
	if err != nil {
		return nil, err
	}
	res = make([]domain.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, r.toDomain(tag))
	}
	er := r.cache.SetBizTags(ctx, biz, bizId, res)
	if er != nil {
		r.l.Error("缓存资源标签失败", logger.Error(er))
	}
	return res, nil
}

func (r *CachedTagRepository) GetBizIdsByTag(ctx context.Context, biz string, tagId int64, offset, limit int, sortBy string) ([]int64, error) {
	return r.dao.GetBizIdsByTag(ctx, biz, tagId, offset, limit, sortBy)
}

func (r *CachedTagRepository) CountBizByTag(ctx context.Context, biz string, tagId int64) (int64, error) {
	return r.dao.CountBizByTag(ctx, biz, tagId)
}

func (r *CachedTagRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Id:          tag.Id,
		Name:        tag.Name,
		Description: tag.Description,
	}
}

func (r *CachedTagRepository) FollowTag(ctx context.Context, uid, tagId int64) error {
	return r.dao.FollowTag(ctx, uid, tagId)
}

func (r *CachedTagRepository) UnfollowTag(ctx context.Context, uid, tagId int64) error {
	return r.dao.UnfollowTag(ctx, uid, tagId)
}

func (r *CachedTagRepository) CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error) {
	return r.dao.CheckTagFollow(ctx, uid, tagId)
}

func (r *CachedTagRepository) GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]domain.Tag, error) {
	tags, err := r.dao.GetUserFollowedTags(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, r.toDomain(tag))
	}
	return res, nil
}

func (r *CachedTagRepository) BatchGetBizTags(ctx context.Context, biz string, bizIds []int64) (map[int64][]domain.Tag, error) {
	tagMap, err := r.dao.BatchGetTagsByBiz(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	result := make(map[int64][]domain.Tag, len(tagMap))
	for bizId, tags := range tagMap {
		domainTags := make([]domain.Tag, 0, len(tags))
		for _, tag := range tags {
			domainTags = append(domainTags, r.toDomain(tag))
		}
		result[bizId] = domainTags
	}
	return result, nil
}

func (r *CachedTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:            tag.Id,
		Name:          tag.Name,
		Description:   tag.Description,
		FollowerCount: tag.FollowerCount,
	}
}
