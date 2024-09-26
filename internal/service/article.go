package service

import (
	"context"
	"webook/internal/domain"
	"webook/internal/pkg/logger"
	"webook/internal/repository/article"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	PublishV1(ctx context.Context, article domain.Article) (int64, error)
	WithDraw(ctx context.Context, article domain.Article) error
}

type articleService struct {
	repo article.ArticleRepository

	// v1 依靠两个不同的 repository 来解决这种跨表，或者夸库的问题
	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository
	l      logger.LoggerV1
}

func NewArticleService(repo article.ArticleRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		repo: repo,
		l:    l,
	}
}

func NewArticleServiceV1(author article.ArticleAuthorRepository, reader article.ArticleReaderRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		author: author,
		reader: reader,
		l:      l,
	}
}

func (s *articleService) WithDraw(ctx context.Context, art domain.Article) error {
	return s.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (s *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	// 制作库
	//id, err := s.repo.Create(ctx, art)
	//// 线上库
	return s.repo.Sync(ctx, art)
}

func (s *articleService) PublishV1(ctx context.Context, article domain.Article) (int64, error) {
	var (
		id  = article.Id
		err error
	)
	if article.Id > 0 {
		err = s.author.Update(ctx, article)
	} else {
		id, err = s.author.Create(ctx, article)
	}
	if err != nil {
		return 0, err
	}

	// 制作库和线上库的 ID 相等
	article.Id = id

	// 对于部分失败，引入重试机制
	for i := 0; i < 3; i++ {
		id, err = s.reader.Save(ctx, article)
		if err == nil {
			break
		}
		s.l.Error("部分失败，保存到线上库失败", logger.Int64("art_id", id), logger.Error(err))
	}
	if err != nil {
		s.l.Error("部分失败，重试彻底失败", logger.Int64("art_id", id), logger.Error(err))
		// 接入告警系统，手工处理一下
		// 走异步，直接保存到本地文件
		// 走 canal
		// 打 MQ
	}
	return id, err
}

func (s *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnPublished
	if art.Id > 0 {
		err := s.repo.Update(ctx, art)
		return art.Id, err
	}
	return s.repo.Create(ctx, art)
}

//func (s *articleService) Update(ctx context.Context, art domain.Article) error {
//	// 只要不更新 author_id
//	// 但是性能较差
//	artInDB := s.repo.FindById(ctx, art.Id)
//	if art.Author.Id != artInDB.Id {
//		return errors.New("更新别人的数据")
//	}
//	return s.repo.Update(ctx, art)
//}
