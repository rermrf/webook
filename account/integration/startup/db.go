package startup

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/account/repository/dao"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/webook_account"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
