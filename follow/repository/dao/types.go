package dao

import "context"

type FollowDao interface {
	// FollowRelationList 获取某人的关注列表
	FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error)
	FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error)
	// CreateFollowRelation 创建联系
	CreateFollowRelation(ctx context.Context, c FollowRelation) error
	UpdateStatus(ctx context.Context, follower int64, followee int64, status uint8) error
	CntFollower(ctx context.Context, uid int64) (int64, error)
	CntFollowee(ctx context.Context, uid int64) (int64, error)
}

type FollowRelation struct {
	Id int64 `gorm:"primaryKey; autoIncrement;"`

	// 在这两个列上，创建一个联合唯一索引
	// 如果你认为查询关注了多少人，是主要查询场景
	// <follower, followee>
	// 如果你认为查询一个人有哪些粉丝，是主要查询场景
	// <followee, follower>
	Follower int64 `gorm:"uniqueIndex:follower_followee"`
	Followee int64 `gorm:"uniqueIndex:follower_follower"`

	Satus uint8

	Ctime int64
	Utime int64
}

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)
