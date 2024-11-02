package repository

import (
	"context"
	"webook/article/domain"
)
//go:generate mockgen -source=./article_author.go -package=repomocks -destination=mocks/article_author_mock.go ArticleAuthorRepository
type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}

type articleAuthorRepository struct{}

func (a articleAuthorRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return 0, nil
}

func (a articleAuthorRepository) Update(ctx context.Context, art domain.Article) error {
	return nil
}

func NewArticleAuthorRepository() ArticleAuthorRepository {
	return &articleAuthorRepository{}
}
