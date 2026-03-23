package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/credit/repository/dao"
	"webook/pkg/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config = Config{
		DSN: "root:root@tcp(localhost:13306)/webook?charset=utf8mb4&parseTime=True&loc=Local",
	}
	err := viper.UnmarshalKey("db.mysql", &cfg)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 初始化表
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}

	// 初始化默认规则
	err = dao.InitDefaultRules(db)
	if err != nil {
		l.Error("初始化默认积分规则失败", logger.Error(err))
	}

	return db
}
