package events

import (
	"context"
	"github.com/IBM/sarama"
	"strconv"
	"time"
	"webook/im/domain"
	"webook/im/service"
	"webook/pkg/canalx"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type MysqlBinlogConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	svc    service.UserSercie
}

func (c *MysqlBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("im_users_sync", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"webook_binlog"},
			// 监听 User 的数据
			saramax.NewHandler[canalx.Message[User]](c.l, c.Consume),
		)
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (c *MysqlBinlogConsumer) Consume(msg *sarama.ConsumerMessage, val canalx.Message[User]) error {
	if val.Table != "users" {
		return nil
	}

	// 删除用户
	if val.Type == "DELETE" {
		// 可以不管
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, data := range val.Data {
		// 将数据同步到 openIM 中
		// 这边可以使用批量接口，看场景，当下不需要
		err := c.svc.Sync(ctx, domain.User{
			Nickname: data.Nickname,
			UserId:   strconv.FormatInt(data.Id, 10),
			//FaceURL:
		})
		if err != nil {
			// 记录日志
			continue
		}
	}
	return nil
}

type User struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`

	// 微信的字段
	WechatUnionID string `json:"wechat_union_id"`
	WechatOpenID  string `json:"wechat_open_id"`
	Nickname      string `json:"nickname"`
	AboutMe       string `json:"about_me"`
	Birthday      int64  `json:"birthday"`
	Ctime         int64  `json:"ctime"`
	Utime         int64  `json:"utime"`
}
