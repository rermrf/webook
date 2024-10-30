package service

import (
	"context"
	"time"
	"webook/internal/domain"
	events "webook/internal/events/article"
	"webook/internal/repository/article"
	logger2 "webook/pkg/logger"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=mocks/article_mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	PublishV1(ctx context.Context, article domain.Article) (int64, error)
	WithDraw(ctx context.Context, article domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
	// ListPub 根据 start 来查询
	ListPub(ctx context.Context, startTime time.Time, offset int, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo article.ArticleRepository

	// v1 依靠两个不同的 repository 来解决这种跨表，或者夸库的问题
	author   article.ArticleAuthorRepository
	reader   article.ArticleReaderRepository
	l        logger2.LoggerV1
	producer events.Producer

	ch chan readInfo
}

type readInfo struct {
	uid int64
	aid int64
}

func NewArticleService(repo article.ArticleRepository, l logger2.LoggerV1, producer events.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		l:        l,
		producer: producer,
		//ch:       make(chan readInfo, 10),
	}
}

func NewArticleServiceV2(repo article.ArticleRepository, l logger2.LoggerV1, producer events.Producer) ArticleService {
	ch := make(chan readInfo, 10)
	go func() {
		uids := make([]int64, 0, 10)
		aids := make([]int64, 0, 10)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		for i := 0; i < 10; i++ {
			select {
			case info, ok := <-ch:
				if !ok {
					cancel()
					return
				}
				uids = append(uids, info.uid)
				aids = append(aids, info.aid)
			case <-ctx.Done():
				break
			}
		}
		cancel()
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		producer.ProducerReadEventV1(ctx, events.ReadEventV1{
			Uids: uids,
			Aids: aids,
		})
		cancel()
	}()
	return &articleService{
		repo:     repo,
		l:        l,
		producer: producer,
		ch:       ch,
	}
}

func NewArticleServiceV1(author article.ArticleAuthorRepository, reader article.ArticleReaderRepository, l logger2.LoggerV1) ArticleService {
	return &articleService{
		author: author,
		reader: reader,
		l:      l,
	}
}

func (s *articleService) ListPub(ctx context.Context, startTime time.Time, offset int, limit int) ([]domain.Article, error) {
	return s.repo.ListPub(ctx, startTime, offset, limit)
}

func (s *articleService) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	art, err := s.repo.GetPublishedById(ctx, id)
	if err == nil {
		go func() {
			// 生产者也可以通过改批量来提高性能
			er := s.producer.ProducerReadEvent(ctx, events.ReadEvent{
				// 即便消费者要用 art 里面的数据
				// 让他去查，而不是在 event 里面带
				Uid: uid,
				Aid: art.Id,
			})
			if er != nil {
				s.l.Error("发送读者阅读时间失败", logger2.Error(err), logger2.Int64("uid", uid), logger2.Int64("artId", art.Id))
				return
			}
		}()

		go func() {
			// 改批量的做法
			s.ch <- readInfo{
				uid: uid,
				aid: art.Id,
			}
		}()
	}

	return art, err
}

func (s *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return s.repo.GetById(ctx, id)
}

func (s *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return s.repo.List(ctx, uid, offset, limit)
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
		s.l.Error("部分失败，保存到线上库失败", logger2.Int64("art_id", id), logger2.Error(err))
	}
	if err != nil {
		s.l.Error("部分失败，重试彻底失败", logger2.Int64("art_id", id), logger2.Error(err))
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
