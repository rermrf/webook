package events

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"webook/notification/domain"
	"webook/notification/service"
	"webook/pkg/logger"
	"webook/pkg/saramax"

	"github.com/IBM/sarama"
)

const TopicNotificationEvents = "notification_events"

// NotificationEventConsumer 通用通知事件消费者
type NotificationEventConsumer struct {
	client sarama.Client
	svc    service.NotificationService
	l      logger.LoggerV1
}

func NewNotificationEventConsumer(
	client sarama.Client,
	svc service.NotificationService,
	l logger.LoggerV1,
) *NotificationEventConsumer {
	return &NotificationEventConsumer{
		client: client,
		svc:    svc,
		l:      l,
	}
}

func (c *NotificationEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("notification", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(
			context.Background(),
			[]string{TopicNotificationEvents},
			saramax.NewHandler[NotificationEvent](c.l, c.Consume),
		)
		if err != nil {
			c.l.Error("退出通知消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (c *NotificationEventConsumer) Consume(msg *sarama.ConsumerMessage, evt NotificationEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	notification := domain.Notification{
		Key:         fmt.Sprintf("notification:%d:%d", evt.SourceId, evt.ReceiverId),
		Channel:     domain.ChannelInApp,
		Receiver:    strconv.FormatInt(evt.ReceiverId, 10),
		GroupType:   domain.NotificationGroupSystem,
		SourceId:    evt.SourceId,
		SourceName:  evt.SourceName,
		TargetId:    evt.TargetId,
		TargetType:  evt.TargetType,
		TargetTitle: evt.TargetTitle,
		Content:     evt.Content,
	}

	_, err := c.svc.Send(ctx, notification)
	return err
}

// LikeEventConsumer 点赞事件消费者
type LikeEventConsumer struct {
	client sarama.Client
	svc    service.NotificationService
	l      logger.LoggerV1
}

func NewLikeEventConsumer(
	client sarama.Client,
	svc service.NotificationService,
	l logger.LoggerV1,
) *LikeEventConsumer {
	return &LikeEventConsumer{
		client: client,
		svc:    svc,
		l:      l,
	}
}

func (c *LikeEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("notification_like", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(
			context.Background(),
			[]string{"like_events"},
			saramax.NewHandler[LikeEvent](c.l, c.Consume),
		)
		if err != nil {
			c.l.Error("退出点赞消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (c *LikeEventConsumer) Consume(msg *sarama.ConsumerMessage, evt LikeEvent) error {
	// 自己点赞自己的不发通知
	if evt.Uid == evt.BizOwnerId {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	notification := domain.Notification{
		Key:         fmt.Sprintf("like:%d:%d", evt.Uid, evt.BizId),
		Channel:     domain.ChannelInApp,
		Receiver:    strconv.FormatInt(evt.BizOwnerId, 10),
		GroupType:   domain.NotificationGroupInteraction,
		SourceId:    evt.Uid,
		SourceName:  evt.UserName,
		TargetId:    evt.BizId,
		TargetType:  evt.Biz,
		TargetTitle: evt.BizTitle,
		Content:     evt.UserName + " 赞了你的" + c.bizTypeToName(evt.Biz),
	}

	_, err := c.svc.Send(ctx, notification)
	return err
}

func (c *LikeEventConsumer) bizTypeToName(biz string) string {
	switch biz {
	case "article":
		return "文章"
	case "comment":
		return "评论"
	default:
		return "内容"
	}
}

// CommentEventConsumer 评论事件消费者
type CommentEventConsumer struct {
	client sarama.Client
	svc    service.NotificationService
	l      logger.LoggerV1
}

func NewCommentEventConsumer(
	client sarama.Client,
	svc service.NotificationService,
	l logger.LoggerV1,
) *CommentEventConsumer {
	return &CommentEventConsumer{
		client: client,
		svc:    svc,
		l:      l,
	}
}

func (c *CommentEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("notification_comment", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(
			context.Background(),
			[]string{"comment_events"},
			saramax.NewHandler[CommentEvent](c.l, c.Consume),
		)
		if err != nil {
			c.l.Error("退出评论消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (c *CommentEventConsumer) Consume(msg *sarama.ConsumerMessage, evt CommentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var notifications []domain.Notification

	// 如果是回复评论，通知被回复的人
	if evt.ParentCommentId > 0 && evt.ParentUserId != evt.Uid {
		notifications = append(notifications, domain.Notification{
			Key:         fmt.Sprintf("comment:reply:%d:%d", evt.Uid, evt.CommentId),
			Channel:     domain.ChannelInApp,
			Receiver:    strconv.FormatInt(evt.ParentUserId, 10),
			GroupType:   domain.NotificationGroupReply,
			SourceId:    evt.Uid,
			SourceName:  evt.UserName,
			TargetId:    evt.CommentId,
			TargetType:  "comment",
			TargetTitle: evt.BizTitle,
			Content:     evt.Content,
		})
	}

	// 通知文章/内容作者（如果评论者不是作者本人，且不是回复作者自己的评论）
	if evt.BizOwnerId != evt.Uid && evt.BizOwnerId != evt.ParentUserId {
		notifications = append(notifications, domain.Notification{
			Key:         fmt.Sprintf("comment:%d:%d", evt.Uid, evt.BizId),
			Channel:     domain.ChannelInApp,
			Receiver:    strconv.FormatInt(evt.BizOwnerId, 10),
			GroupType:   domain.NotificationGroupReply,
			SourceId:    evt.Uid,
			SourceName:  evt.UserName,
			TargetId:    evt.BizId,
			TargetType:  evt.Biz,
			TargetTitle: evt.BizTitle,
			Content:     evt.Content,
		})
	}

	if len(notifications) == 0 {
		return nil
	}

	_, err := c.svc.BatchSend(ctx, notifications)
	return err
}

// FollowEventConsumer 关注事件消费者
type FollowEventConsumer struct {
	client sarama.Client
	svc    service.NotificationService
	l      logger.LoggerV1
}

func NewFollowEventConsumer(
	client sarama.Client,
	svc service.NotificationService,
	l logger.LoggerV1,
) *FollowEventConsumer {
	return &FollowEventConsumer{
		client: client,
		svc:    svc,
		l:      l,
	}
}

func (c *FollowEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("notification_follow", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(
			context.Background(),
			[]string{"follow_events"},
			saramax.NewHandler[FollowEvent](c.l, c.Consume),
		)
		if err != nil {
			c.l.Error("退出关注消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (c *FollowEventConsumer) Consume(msg *sarama.ConsumerMessage, evt FollowEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	notification := domain.Notification{
		Key:        fmt.Sprintf("follow:%d:%d", evt.FollowerId, evt.FolloweeId),
		Channel:    domain.ChannelInApp,
		Receiver:   strconv.FormatInt(evt.FolloweeId, 10),
		GroupType:  domain.NotificationGroupFollow,
		SourceId:   evt.FollowerId,
		SourceName: evt.FollowerName,
		Content:    evt.FollowerName + " 关注了你",
	}

	_, err := c.svc.Send(ctx, notification)
	return err
}

// CollectEventConsumer 收藏事件消费者
type CollectEventConsumer struct {
	client sarama.Client
	svc    service.NotificationService
	l      logger.LoggerV1
}

func NewCollectEventConsumer(
	client sarama.Client,
	svc service.NotificationService,
	l logger.LoggerV1,
) *CollectEventConsumer {
	return &CollectEventConsumer{
		client: client,
		svc:    svc,
		l:      l,
	}
}

func (c *CollectEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("notification_collect", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(
			context.Background(),
			[]string{"collect_events"},
			saramax.NewHandler[CollectEvent](c.l, c.Consume),
		)
		if err != nil {
			c.l.Error("退出收藏消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (c *CollectEventConsumer) Consume(msg *sarama.ConsumerMessage, evt CollectEvent) error {
	// 自己收藏自己的不发通知
	if evt.Uid == evt.BizOwnerId {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	notification := domain.Notification{
		Key:         fmt.Sprintf("collect:%d:%d", evt.Uid, evt.BizId),
		Channel:     domain.ChannelInApp,
		Receiver:    strconv.FormatInt(evt.BizOwnerId, 10),
		GroupType:   domain.NotificationGroupInteraction,
		SourceId:    evt.Uid,
		SourceName:  evt.UserName,
		TargetId:    evt.BizId,
		TargetType:  evt.Biz,
		TargetTitle: evt.BizTitle,
		Content:     evt.UserName + " 收藏了你的" + c.bizTypeToName(evt.Biz),
	}

	_, err := c.svc.Send(ctx, notification)
	return err
}

func (c *CollectEventConsumer) bizTypeToName(biz string) string {
	switch biz {
	case "article":
		return "文章"
	default:
		return "内容"
	}
}
