package article

import (
	"context"
	"webook/internal/domain"
)

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
