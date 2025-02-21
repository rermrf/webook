package domain

type FollowRelation struct {
	// 被关注的人
	Followee int64
	// 关注的人
	Follower int64
	// 根据你的业务需求，可以在这里加字段
	// 比如说备注啊，标签啥的
	// Note string
}

type FollowStatics struct {
	// 被多少人关注
	Followers int64
	// 自己关注了多少人
	Followees int64
}
