package article

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/pkg/logger"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	dao "webook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// Sync 存储并同步数据
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
	//FindById(ctx context.Context, id int64) domain.Article
}

type CachedArticleRepository struct {
	dao      dao.ArticleDao
	userRepo repository.UserRepository
	cache    cache.ArticleCache
	l        logger.LoggerV1
}

func NewArticleRepository(dao dao.ArticleDao, cache cache.ArticleCache, l logger.LoggerV1) ArticleRepository {
	return &CachedArticleRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (r *CachedArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	// 读取线上库数据，如果你的 Content 被你放过去了 OSS 上， 就要让前端去读 Content 字段
	art, err := r.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 这边要组装 user 了， 适合单体应用
	user, err := r.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}
	res := domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticleStatus(art.Status),
		Author: domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}

	return res, nil
}

func (r *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := r.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := r.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = r.toDomain(art)
	go func() {
		er := r.cache.Set(ctx, id, res)
		if er != nil {
			// 记录日志
			r.l.Error("缓存设置错误", logger.Error(err))
		}
	}()
	return res, nil
}

func (r *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 在这个地方，集成你的复杂的缓存方案
	// 一般来说，如果 offset、limit、authorId 发生变化缓存就无效了
	// 所以我们只缓存第一页
	if offset == 0 && limit <= 100 {
		data, err := r.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				r.preCache(ctx, data)
			}()
			//return data[:limit], nil
			return data, nil
		}
	}
	res, err := r.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := r.toDomains(res)
	// 回写缓存的时候，可以同步，也可以异步
	go func() {
		err := r.cache.SetFirstPage(ctx, uid, data)
		r.l.Error("回写缓存失败", logger.Error(err))
		// 缓存预加载
		// 用户在查看自己的文章列表时大概率会查看自己的最新发布或修改的文章
		r.preCache(ctx, data)
	}()
	return r.toDomains(res), err
}

func (r *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error {
	defer func() {
		// 清空缓存
		err := r.cache.DelFirstPage(ctx, author)
		if err != nil {
			r.l.Error("清除缓存失败", logger.Error(err))
		}
	}()
	return r.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (r *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := r.dao.Sync(ctx, r.toEntity(art))
	if err == nil {
		// 新增或修改成功，清空缓存
		er := r.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 不需要特别关心
			// 打印 WARN 日志
			return 0, er
		}
		// 提前缓存好线上库数据
		er = r.cache.SetPub(ctx, art)
		if er != nil {
			return 0, er
		}
	}
	return id, err
}

func (r *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		err := r.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			r.l.Error("清除缓存失败", logger.Error(err))
		}
	}()
	return r.dao.Insert(ctx, r.toEntity(art))
}

func (r *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		// 清空缓存
		err := r.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			r.l.Error("清除缓存失败", logger.Error(err))
		}
	}()
	return r.dao.UpdateById(ctx, r.toEntity(art))
}

func (r *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (r *CachedArticleRepository) preCache(ctx context.Context, data []domain.Article) {
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := r.cache.Set(ctx, data[0].Id, data[0])
		if err != nil {
			r.l.Error("提前预加载缓存失败", logger.Error(err))
		}
	}
}

func (r *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticleStatus(art.Status),
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
}

func (r *CachedArticleRepository) toDomains(arts []dao.Article) []domain.Article {
	res := make([]domain.Article, len(arts))
	for k, v := range arts {
		res[k] = r.toDomain(v)
	}
	return res
}

//type CachedArticleRepository struct {
//	dao dao.ArticleDao
//
//	// V1 操作两个 DAO
//	readerDAO dao.ReaderDAO
//	authorDAO dao.AuthorDAO
//
//	// 耦合了 DAO 操作的东西
//	// 正常情况下，如果要在 repo 层面上操作事物
//	// 那么就只能利用 db 开始事物之后，创建基于事物的 DAO
//	// 或者，直接去掉 DAO 这一层，在 repo 的实现上，直接操作数据库
//	db *gorm.DB
//}

// SyncV2 尝试在 repo 层面上解决事物问题
// 确保保存到制作库和线上库同时成功或同时失败
//func (r *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
//	// 开启一个事物
//	tx := r.db.WithContext(ctx).Begin()
//	if tx.Error != nil {
//		return 0, tx.Error
//	}
//	// 防止在 commit 或者 rollback 之前发生panic，导致事物一直挂在数据库
//	defer tx.Rollback()
//	// 利用 tx 来构建 DAO
//	author := dao.NewAuthorDAO(tx)
//	reader := dao.NewReaderDAO(tx)
//
//	var (
//		id  = art.Id
//		err error
//	)
//	artn := r.toEntity(art)
//	// 应该先保存到制作库，再保存到线上库
//	if id > 0 {
//		err = author.UpdateById(ctx, artn)
//	} else {
//		id, err = author.Insert(ctx, artn)
//	}
//	if err != nil {
//		// 回滚事物
//		//tx.Rollback()
//		return 0, err
//	}
//	// 操作线上库，保存数据，同步过来
//	// 考虑到，此时线上库可能有，可能没有，要有一个 UPSERT 的写法
//	// INSERT or UPDATE
//	// 如果数据库有，那么就更新，不然就插入
//	err = reader.UpsertV2(ctx, dao.PublishedArticle{Article: artn})
//	// 执行成功直接提交
//	tx.Commit()
//	return id, err
//}
//
//func (r *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
//	var (
//		id  = art.Id
//		err error
//	)
//	artn := r.toEntity(art)
//	// 应该先保存到制作库，再保存到线上库
//	if id > 0 {
//		err = r.authorDAO.UpdateById(ctx, artn)
//	} else {
//		id, err = r.authorDAO.Insert(ctx, artn)
//	}
//	if err != nil {
//		return 0, err
//	}
//	// 操作线上库，保存数据，同步过来
//	// 考虑到，此时线上库可能有，可能没有，要有一个 UPSERT 的写法
//	// INSERT or UPDATE
//	// 如果数据库有，那么就更新，不然就插入
//	err = r.readerDAO.Upsert(ctx, artn)
//	return id, err
//}
