package events

import (
	"context"
	"github.com/IBM/sarama"
	"strconv"
	"time"
	"webook/feed/domain"
	"webook/feed/service"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

const topicArticleEvent = "article_feed_event"

// ArticleFeedEvent 由业务方定义，本服务做适配
type ArticleFeedEvent struct {
	uid int64
	aid int64
}

type ArticleEventConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	svc    service.FeedService
}

func NewArticleEventConsumer(client sarama.Client, l logger.LoggerV1, svc service.FeedService) *ArticleEventConsumer {
	return &ArticleEventConsumer{client: client, l: l, svc: svc}
}

func (a *ArticleEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("articleFeed", a.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{topicArticleEvent}, saramax.NewHandler[ArticleFeedEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出了消费循环", logger.Error(err))
		}
	}()
	return err
}

func (a *ArticleEventConsumer) Consume(msg *sarama.ConsumerMessage, t ArticleFeedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return a.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: service.FollowEventName,
		Ext: map[string]string{
			"uid": strconv.FormatInt(t.uid, 10),
			"aid": strconv.FormatInt(t.aid, 10),
		},
	})
}
