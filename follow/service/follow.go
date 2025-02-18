package service

import (
	"context"
	"webook/follow/domain"
	"webook/follow/repository"
)

type followService struct {
	repo repository.FollowRepository
}

func NewFollowService(repo repository.FollowRepository) FollowService {
	return &followService{repo: repo}
}

func (f *followService) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	return f.repo.GetFollowee(ctx, follower, offset, limit)
}

func (f *followService) FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error) {
	return f.repo.FollowInfo(ctx, follower, followee)
}

func (f *followService) Follow(ctx context.Context, follower, followee int64) error {
	return f.repo.AddFollowRelation(ctx, domain.FollowRelation{
		Follower: follower,
		Followee: followee,
	})
}

func (f *followService) CancelFollow(ctx context.Context, follower, followee int64) error {
	return f.repo.InactiveFollowRelation(ctx, follower, followee)
}

func (f *followService) GetFollower(ctx context.Context, followee, offset, limit int64) ([]domain.FollowRelation, error) {
	return f.repo.GetFollower(ctx, followee, offset, limit)
}

func (f *followService) GetFollowStatic(ctx context.Context, followee int64) (domain.FollowStatics, error) {
	return f.repo.GetFollowStatics(ctx, followee)
}
