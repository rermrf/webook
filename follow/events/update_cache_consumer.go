package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/follow/repository"
	"webook/follow/repository/dao"
	"webook/pkg/canalx"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

// MysqlBinlogConsumer 通过 canal 监听mysql变化更新缓存
type MysqlBinlogConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	repo   *repository.CachedFollowRepository
}

func (c *MysqlBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("follow_relation_cache", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"webook_binlog"}, saramax.NewHandler[canalx.Message[dao.FollowRelation]](c.l, c.Consume))
		if err != nil {
			// 记录日志
			c.l.Error("退出了消费循环", logger.Error(err))
		}
	}()
	return err
}

func (c *MysqlBinlogConsumer) Consume(msg *sarama.ConsumerMessage, val canalx.Message[dao.FollowRelation]) error {
	if val.Table != "follow_relations" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, data := range val.Data {
		var err error
		switch data.Satus {
		case uint8(dao.FollowRelationStatusActive):
			err = c.repo.Cache().Follow(ctx, data.Follower, data.Followee)
		case uint8(dao.FollowRelationStatusInactive):
			err = c.repo.Cache().CancelFollow(ctx, data.Follower, data.Followee)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
