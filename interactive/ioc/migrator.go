package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"webook/interactive/repository/dao"
	"webook/pkg/ginx"
	"webook/pkg/gormx/connpool"
	"webook/pkg/logger"
	"webook/pkg/migrator/events"
	"webook/pkg/migrator/events/fixer"
	"webook/pkg/migrator/scheduler"
)

const topic = "migrator_interactive"

func InitMigratorServer(src SrcDB, dst DstDB, pool *connpool.DoubleWritePool, l logger.LoggerV1, producer events.Producer) *ginx.Server {
	// 在这里，有多少张表，就初始化多少个 scheduler
	intrSch := scheduler.NewScheduler[dao.Interactive](src, dst, pool, l, producer)
	addr := viper.GetString("migrator.web.addr")
	engine := gin.Default()
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "emoji",
		Subsystem: "webook_intr_admin",
		Name:      "http_biz_code",
		Help:      "统计http",
	})
	intrSch.RegisterRoutes(engine.Group("/migrator/interactive"))
	return &ginx.Server{
		Addr:   addr,
		Engine: engine,
	}
}

func InitMigradatorProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(p, topic)
}

func InitFixDataConsumer(l logger.LoggerV1, src SrcDB, dst DstDB, client sarama.Client) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, topic, src, dst)
	if err != nil {
		panic(err)
	}
	return res
}
