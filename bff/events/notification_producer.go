package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

// NotificationProducer 通知事件生产者
type NotificationProducer interface {
	// ProduceLikeEvent 发送点赞通知事件
	ProduceLikeEvent(ctx context.Context, evt LikeEvent) error
	// ProduceCollectEvent 发送收藏通知事件
	ProduceCollectEvent(ctx context.Context, evt CollectEvent) error
	// ProduceCommentEvent 发送评论通知事件
	ProduceCommentEvent(ctx context.Context, evt CommentEvent) error
	// ProduceFollowEvent 发送关注通知事件
	ProduceFollowEvent(ctx context.Context, evt FollowEvent) error
}

// LikeEvent 点赞事件
type LikeEvent struct {
	Uid        int64  `json:"uid"`         // 点赞者ID
	UserName   string `json:"user_name"`   // 点赞者名称
	Biz        string `json:"biz"`         // 业务类型 (article/comment)
	BizId      int64  `json:"biz_id"`      // 业务对象ID
	BizTitle   string `json:"biz_title"`   // 业务对象标题
	BizOwnerId int64  `json:"biz_owner_id"` // 作者ID
}

// CollectEvent 收藏事件
type CollectEvent struct {
	Uid        int64  `json:"uid"`
	UserName   string `json:"user_name"`
	Biz        string `json:"biz"`
	BizId      int64  `json:"biz_id"`
	BizTitle   string `json:"biz_title"`
	BizOwnerId int64  `json:"biz_owner_id"`
}

// CommentEvent 评论事件
type CommentEvent struct {
	CommentId       int64  `json:"comment_id"`
	Uid             int64  `json:"uid"`
	UserName        string `json:"user_name"`
	Biz             string `json:"biz"`
	BizId           int64  `json:"biz_id"`
	BizTitle        string `json:"biz_title"`
	BizOwnerId      int64  `json:"biz_owner_id"`
	Content         string `json:"content"`
	ParentCommentId int64  `json:"parent_comment_id"`
	ParentUserId    int64  `json:"parent_user_id"`
}

// FollowEvent 关注事件
type FollowEvent struct {
	FollowerId   int64  `json:"follower_id"`
	FollowerName string `json:"follower_name"`
	FolloweeId   int64  `json:"followee_id"`
}

type SaramaNotificationProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaNotificationProducer(producer sarama.SyncProducer) NotificationProducer {
	return &SaramaNotificationProducer{producer: producer}
}

func (p *SaramaNotificationProducer) ProduceLikeEvent(ctx context.Context, evt LikeEvent) error {
	return p.send("like_events", evt)
}

func (p *SaramaNotificationProducer) ProduceCollectEvent(ctx context.Context, evt CollectEvent) error {
	return p.send("collect_events", evt)
}

func (p *SaramaNotificationProducer) ProduceCommentEvent(ctx context.Context, evt CommentEvent) error {
	return p.send("comment_events", evt)
}

func (p *SaramaNotificationProducer) ProduceFollowEvent(ctx context.Context, evt FollowEvent) error {
	return p.send("follow_events", evt)
}

func (p *SaramaNotificationProducer) send(topic string, evt any) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	})
	return err
}
