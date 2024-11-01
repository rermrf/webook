package dao

import (
	"gorm.io/gorm"
	"webook/internal/repository/dao/article"
	"webook/user/repository/dao"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&dao.User{}, &article.Article{},
		&article.PublishedArticle{},
		&Job{},
	)
}
