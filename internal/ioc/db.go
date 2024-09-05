package ioc

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/config"
	"webook/internal/repository/dao"
)

func InitDB() *gorm.DB {
	// 使用 k8s 部署的 mysql
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic("failed to connect database")
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
