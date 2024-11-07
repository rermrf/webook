package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
	"webook/interactive/repository/dao"
	"webook/pkg/gormx"
	"webook/pkg/gormx/connpool"
	"webook/pkg/logger"
)

func InitSRC(l logger.LoggerV1) SrcDB {
	return InitDB(l, "src")
}

func InitDST(l logger.LoggerV1) DstDB {
	return InitDB(l, "dst")
}

func InitDoubleWritePool(src SrcDB, dst DstDB) *connpool.DoubleWritePool {
	pattern := viper.GetString("migrator.pattern")
	return connpool.NewDoubleWritePool(src.ConnPool, dst.ConnPool, pattern)
}

// InitBizDB 业务用的，支持双写的DB
func InitBizDB(connPool *connpool.DoubleWritePool) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: connPool,
	}))
	if err != nil {
		panic(err)
	}
	return db
}

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitDB(l logger.LoggerV1, key string) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config = Config{
		// 这只默认值
		DSN: "root:root@tcp(localhost:3306)/webook?charset=utf8mb4&parseTime=True&loc=Local",
	}
	err := viper.UnmarshalKey("db.mysql."+key, &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		//Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
		//	SlowThreshold:             time.Millisecond * 100,
		//	IgnoreRecordNotFoundError: true,
		//	ParameterizedQueries:      false,
		//	LogLevel:                  glogger.Info,
		//}),
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

	// 监控查询的执行时间
	pcb := gormx.NewPrometheusSummaryMetricPlugin(
		"emoji", "webook", "gorm_query_time"+key, "统计 GORM 的执行时间", "webook")
	err = db.Use(pcb)
	if err != nil {
		panic(err)
	}

	db.Use(tracing.NewPlugin(tracing.WithDBName("webook"),
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
