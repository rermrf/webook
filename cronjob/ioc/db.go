package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
	"time"
	"webook/cronjob/repository/dao"
	"webook/pkg/gormx"
	"webook/pkg/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config = Config{
		// 这只默认值
		DSN: "root:root@tcp(localhost:3306)/webook?charset=utf8mb4&parseTime=True&loc=Local",
	}
	err := viper.UnmarshalKey("db.mysql", &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			// 满查询阈值，只有执行时间超过这个阈值，才会使用
			// 50ms, 100ms
			// SQL 查询必然要求命中索引，最好就是走一次磁盘 IO
			// 一次磁盘 IO 是不到 10ms
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			LogLevel:                  glogger.Info,
		}),
	})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	prom := gormx.NewPrometheusSummaryMetricPlugin("emoji", "webook", "gorm_query_time", "统计 GORM 的执行时间", "webook")
	err = prom.Initialize(db)
	if err != nil {
		panic(err)
	}

	// otel 记录
	err = db.Use(tracing.NewPlugin(tracing.WithDBName("webook"),
		tracing.WithQueryFormatter(
			func(query string) string {
				l.Debug("", logger.String("query", query))
				return query
			}),
		// 不要记录 metrics 部分
		tracing.WithoutMetrics(),
		// 不要记录查询参数
		tracing.WithoutQueryVariables(),
	))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}

	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}
