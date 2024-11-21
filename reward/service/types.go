package service

import (
	"context"
	"webook/reward/domain"
)

//go:generate mockgen -source=./types.go -destination=mocks/reward_mock.go -package=svcmocks RewardService
type RewardService interface {
	// PreReward 准备打赏
	PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error)
	UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error
}
