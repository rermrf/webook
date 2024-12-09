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

const topicArticleEvent = "article_published_event"

// ArticlePublishedEvent 由业务方定义，本服务做适配
// 监听业务方的事件
// 业务方强势，我们自己适配
type ArticlePublishedEvent struct {
	Uid int64
	Aid int64
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
		err := cg.Consume(context.Background(), []string{topicArticleEvent}, saramax.NewHandler[ArticlePublishedEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出了消费循环", logger.Error(err))
		}
	}()
	return err
}

func (a *ArticleEventConsumer) Consume(msg *sarama.ConsumerMessage, t ArticlePublishedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return a.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: service.FollowEventName,
		Ext: map[string]string{
			"uid": strconv.FormatInt(t.Uid, 10),
			"aid": strconv.FormatInt(t.Aid, 10),
		},
	})
}
