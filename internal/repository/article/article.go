package article

import (
	"context"
	"webook/internal/domain"
	dao "webook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// Sync 存储并同步数据
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error
	//FindById(ctx context.Context, id int64) domain.Article
}

type CachedArticleRepository struct {
	dao dao.ArticleDao
}

func NewArticleRepository(dao dao.ArticleDao) *CachedArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func (r *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error {
	return r.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (r *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return r.dao.Sync(ctx, r.toEntity(art))
}

func (r *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return r.dao.Insert(ctx, r.toEntity(art))
}

func (r *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
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
//	err = reader.UpsertV2(ctx, dao.PublishArticle{Article: artn})
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
