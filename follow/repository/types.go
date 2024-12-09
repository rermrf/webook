package repository

import (
	"context"
	"webook/follow/domain"
)

type FollowRepository interface {
	// GetFollowee 获取某人和关注列表
	GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error)
	// FollowInfo 查看关注人的详情
	FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error)
	// AddFollowRelation 创建关注关系
	AddFollowRelation(ctx context.Context, f domain.FollowRelation) error
	// InactiveFollowRelation 取消关注
	InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error
	GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error)
	// GetFollower 获取某人粉丝
	GetFollower(ctx context.Context, followee int64) ([]domain.FollowRelation, error)
}
