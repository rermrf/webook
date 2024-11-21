package repository

import (
	"context"
	"webook/reward/domain"
	"webook/reward/repository/cache"
	"webook/reward/repository/dao"
)

type rewardRepository struct {
	dao   dao.RewardDAO
	cache cache.RewardCache
}

func NewRewardRepository(dao dao.RewardDAO, cache cache.RewardCache) RewardRepository {
	return &rewardRepository{dao: dao, cache: cache}
}

func (repo *rewardRepository) CreateReward(ctx context.Context, reward domain.Reward) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(reward))
}

func (repo *rewardRepository) GetReward(ctx context.Context, rid int64) (domain.Reward, error) {
	r, err := repo.dao.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	return repo.toDomain(r), nil
}

func (repo *rewardRepository) GetCachedCodeURL(ctx context.Context, reward domain.Reward) (domain.CodeURL, error) {
	return repo.cache.GetCachedCodeURL(ctx, reward)
}

func (repo *rewardRepository) CachedCodeURL(ctx context.Context, cu domain.CodeURL, reward domain.Reward) error {
	return repo.cache.CacheCodeURL(ctx, cu, reward)
}

func (repo *rewardRepository) UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error {
	return repo.dao.UpdateStatus(ctx, rid, status.AsUint8())
}

func (repo *rewardRepository) toEntity(reward domain.Reward) dao.Reward {
	return dao.Reward{
		Biz:     reward.Target.Biz,
		BizId:   reward.Target.BizId,
		BizName: reward.Target.BizName,
		Uid:     reward.Uid,
		Amount:  reward.Amt,
		Status:  reward.Status.AsUint8(),
	}
}

func (repo *rewardRepository) toDomain(r dao.Reward) domain.Reward {
	return domain.Reward{
		Id: r.Id,
		Target: domain.Target{
			Biz:     r.Biz,
			BizId:   r.BizId,
			BizName: r.BizName,
			Uid:     r.Uid,
		},
		Amt:    r.Amount,
		Status: domain.RewardStatus(r.Status),
	}
}
