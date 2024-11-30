package repository

import (
	"context"
	"webook/search/domain"
	"webook/search/repository/dao"
)

type articleRepository struct {
	dao  dao.ArticleDao
	tags dao.TagDao
}

func NewArticleRepository(dao dao.ArticleDao, tags dao.TagDao) ArticleRepository {
	return &articleRepository{dao: dao, tags: tags}
}

func (a *articleRepository) InputArticle(ctx context.Context, msg domain.Article) error {
	return a.dao.InputArticle(ctx, dao.Article{
		Id:      msg.Id,
		Title:   msg.Title,
		Status:  msg.Status,
		Content: msg.Content,
	})
}

func (a *articleRepository) SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error) {
	ids, err := a.tags.Search(ctx, uid, "article", keywords)
	if err != nil {
		return nil, err
	}
	arts, err := a.dao.Search(ctx, ids, keywords)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Article, 0, len(arts))
	for _, art := range arts {
		res = append(res, domain.Article{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status,
			Content: art.Content,
			//Tags: art.Tags,
		})
	}
	return res, nil

}
