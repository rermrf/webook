package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/article/domain"
	"webook/article/repository"
	"webook/article/repository/dao"
	"webook/pkg/canalx"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type MysqlBinlogConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	//耦合到实现，而不是耦合到接口，除非你把操作缓存的方法，也定义到 repository 接口上
	repo *repository.CachedArticleRepository
}

func (m *MysqlBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("public_article_cache", m.client)
	if err != nil {
		return err
	}
	go func() {
		// 这里逼不得已和 DAO 耦合在一起
		err := cg.Consume(context.Background(), []string{"webook_binlog"}, saramax.NewHandler[canalx.Message[dao.PublishedArticle]](m.l, m.Consume))
		if err != nil {
			m.l.Error("退出了消费循环", logger.Error(err))
		}
	}()
	return err
}

func (m *MysqlBinlogConsumer) Consume(msg *sarama.ConsumerMessage, val canalx.Message[dao.PublishedArticle]) error {
	// 别的表的 binlog，不需要关心
	// 可以考虑不同的表用不同的 topic ，那么这里就不需要判定了
	if val.Table != "published_articles" {
		return nil
	}

	// 更新缓存
	// 增删改的消息，实际上在 publish article 里面是没有删的消息
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, data := range val.Data {
		var err error
		switch data.Status {
		case domain.ArticleStatusPublished.ToUint8():
			// 发表要写入
			err = m.repo.Cache().SetPub(ctx, m.repo.ToDomain(dao.Article(data)))
		case domain.ArticleStatusPrivate.ToUint8():
			err = m.repo.Cache().DelPub(ctx, data.Id)
		}
		if err != nil {
			// 记录日志就行
			m.l.Error("使用canal通知缓存修改失败", logger.Error(err))
		}
	}
	return nil
}
