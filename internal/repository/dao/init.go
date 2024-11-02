package dao

import (
	"gorm.io/gorm"
	"webook/article/repository/dao"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&dao.Article{},
		&dao.PublishedArticle{},
		&Job{},
	)
}
