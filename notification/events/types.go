package events

// NotificationEvent 通知事件（用于Kafka消息传递）
type NotificationEvent struct {
	// Type 通知类型: like/collect/comment/reply/follow/mention/feed/system
	Type string `json:"type"`
	// SourceId 触发者ID
	SourceId int64 `json:"source_id"`
	// SourceName 触发者名称
	SourceName string `json:"source_name"`
	// TargetId 目标对象ID（文章ID/评论ID等）
	TargetId int64 `json:"target_id"`
	// TargetType 目标对象类型（article/comment等）
	TargetType string `json:"target_type"`
	// TargetTitle 目标对象标题/摘要
	TargetTitle string `json:"target_title"`
	// ReceiverId 接收者ID
	ReceiverId int64 `json:"receiver_id"`
	// Content 通知内容（可选）
	Content string `json:"content"`
}

// LikeEvent 点赞事件
type LikeEvent struct {
	// Uid 点赞者ID
	Uid int64 `json:"uid"`
	// UserName 点赞者名称
	UserName string `json:"user_name"`
	// Biz 业务类型 (article/comment等)
	Biz string `json:"biz"`
	// BizId 业务对象ID
	BizId int64 `json:"biz_id"`
	// BizTitle 业务对象标题
	BizTitle string `json:"biz_title"`
	// BizOwnerId 业务对象拥有者ID（文章作者、评论者等）
	BizOwnerId int64 `json:"biz_owner_id"`
}

// CollectEvent 收藏事件
type CollectEvent struct {
	// Uid 收藏者ID
	Uid int64 `json:"uid"`
	// UserName 收藏者名称
	UserName string `json:"user_name"`
	// Biz 业务类型 (article等)
	Biz string `json:"biz"`
	// BizId 业务对象ID
	BizId int64 `json:"biz_id"`
	// BizTitle 业务对象标题
	BizTitle string `json:"biz_title"`
	// BizOwnerId 业务对象拥有者ID
	BizOwnerId int64 `json:"biz_owner_id"`
}

// CommentEvent 评论事件
type CommentEvent struct {
	// CommentId 评论ID
	CommentId int64 `json:"comment_id"`
	// Uid 评论者ID
	Uid int64 `json:"uid"`
	// UserName 评论者名称
	UserName string `json:"user_name"`
	// Biz 业务类型
	Biz string `json:"biz"`
	// BizId 业务对象ID
	BizId int64 `json:"biz_id"`
	// BizTitle 业务对象标题
	BizTitle string `json:"biz_title"`
	// BizOwnerId 业务对象拥有者ID
	BizOwnerId int64 `json:"biz_owner_id"`
	// Content 评论内容摘要
	Content string `json:"content"`
	// ParentCommentId 父评论ID（回复时有值）
	ParentCommentId int64 `json:"parent_comment_id"`
	// ParentUserId 父评论作者ID（回复时有值）
	ParentUserId int64 `json:"parent_user_id"`
}

// FollowEvent 关注事件
type FollowEvent struct {
	// FollowerId 关注者ID
	FollowerId int64 `json:"follower_id"`
	// FollowerName 关注者名称
	FollowerName string `json:"follower_name"`
	// FolloweeId 被关注者ID
	FolloweeId int64 `json:"followee_id"`
}

// FeedEvent Feed更新事件
type FeedEvent struct {
	// AuthorId 作者ID
	AuthorId int64 `json:"author_id"`
	// AuthorName 作者名称
	AuthorName string `json:"author_name"`
	// ArticleId 文章ID
	ArticleId int64 `json:"article_id"`
	// ArticleTitle 文章标题
	ArticleTitle string `json:"article_title"`
	// FollowerIds 粉丝ID列表（需要通知的用户）
	FollowerIds []int64 `json:"follower_ids"`
}
