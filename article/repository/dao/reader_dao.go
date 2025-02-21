package dao

import (
	"context"
	"gorm.io/gorm"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, art Article) error
	UpsertV2(ctx context.Context, art PublishedArticle) error
}

func NewReaderDAO(db *gorm.DB) ReaderDAO {
	panic("implement me")
}

// PublishedArticle 这个代表的是线上表
//type PublishedArticle struct {
//	Article
//}
