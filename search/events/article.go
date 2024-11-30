package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/pkg/logger"
	"webook/pkg/saramax"
	"webook/search/domain"
	"webook/search/service"
)

const topicSyncArticle = "sync_article_event"

type ArticleConsumer struct {
	syncSvc service.SyncService
	client  sarama.Client
	l       logger.LoggerV1
}

func NewArticleConsumer(syncSvc service.SyncService, client sarama.Client, l logger.LoggerV1) *ArticleConsumer {
	return &ArticleConsumer{syncSvc: syncSvc, client: client, l: l}
}

type ArticleEvent struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Status  int32  `json:"status"`
	Content string `json:"content"`
}

func (a *ArticleConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("sync_article", a.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{topicSyncArticle}, saramax.NewHandler[ArticleEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (a *ArticleConsumer) Consume(msg *sarama.ConsumerMessage, evt ArticleEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return a.syncSvc.InputArticle(ctx, a.toDomain(evt))
}

func (a *ArticleConsumer) toDomain(article ArticleEvent) domain.Article {
	return domain.Article{
		Id:      article.Id,
		Title:   article.Title,
		Status:  article.Status,
		Content: article.Content,
	}
}
