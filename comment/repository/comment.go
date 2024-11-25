package repository

import (
	"context"
	"database/sql"
	"golang.org/x/sync/errgroup"
	"time"
	"webook/comment/domain"
	"webook/comment/repository/dao"
	"webook/pkg/logger"
)

type commentRepository struct {
	dao dao.CommentDao
	l   logger.LoggerV1
}

func NewCommentRepository(dao dao.CommentDao, l logger.LoggerV1) CommentRepository {
	return &commentRepository{
		dao: dao,
		l:   l,
	}
}

func (c *commentRepository) FindByBiz(ctx context.Context, biz string, bizId, minID, limit int64) ([]domain.Comment, error) {
	conmments, err := c.dao.FindByBiz(ctx, biz, bizId, minID, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(conmments))
	// 只找三条
	var eg errgroup.Group
	downgraded := ctx.Value("downgraded") == "true"
	for _, conmment := range conmments {
		conmment := conmment
		// 这两句不能放进去，因为并发操作 res 会有坑
		cm := c.toDomain(conmment)
		res = append(res, cm)
		if downgraded {
			continue
		}
		eg.Go(func() error {
			// 只展示三条
			cm.Children = make([]domain.Comment, 0, 3)
			rs, err := c.dao.FindRepliesByPid(ctx, conmment.Id, 0, 3)
			if err != nil {
				// 我们认为这个错误是可以容忍的
				c.l.Error("查询子评论失败", logger.Error(err))
				return nil
			}
			for _, r := range rs {
				cm.Children = append(cm.Children, c.toDomain(r))
			}
			return nil
		})
	}
	return res, eg.Wait()
}

func (c *commentRepository) DeleteComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Delete(ctx, c.toEntity(comment))
}

func (c *commentRepository) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Insert(ctx, c.toEntity(comment))
}

func (c *commentRepository) GetCommentByIds(ctx context.Context, ids []int64) ([]domain.Comment, error) {
	vals, err := c.dao.FindOneByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	comments := make([]domain.Comment, 0, len(vals))
	for _, v := range vals {
		comment := c.toDomain(v)
		comments = append(comments, comment)
	}
	return comments, nil
}

func (c *commentRepository) GetMoreReplies(ctx context.Context, rid int64, maxId int64, limit int64) ([]domain.Comment, error) {
	cs, err := c.dao.FindRepliesByRid(ctx, rid, maxId, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(cs))
	for _, cm := range cs {
		res = append(res, c.toDomain(cm))
	}
	return res, nil
}

func (c *commentRepository) toDomain(src dao.Comment) domain.Comment {
	res := domain.Comment{
		Id: src.Id,
		Commentator: domain.User{
			ID: src.Uid,
		},
		Biz:     src.Biz,
		BizId:   src.BizId,
		Content: src.Content,
		Ctime:   time.UnixMilli(src.Ctime),
		Utime:   time.UnixMilli(src.Utime),
	}
	if src.PID.Valid {
		res.ParentComment = &domain.Comment{
			Id: src.PID.Int64,
		}
	}
	if src.RootId.Valid {
		res.RootComment = &domain.Comment{
			Id: src.RootId.Int64,
		}
	}
	return res
}

func (c *commentRepository) toEntity(src domain.Comment) dao.Comment {
	res := dao.Comment{
		Id:      src.Id,
		Uid:     src.Commentator.ID,
		Biz:     src.Biz,
		BizId:   src.BizId,
		Content: src.Content,
	}
	if src.RootComment != nil {
		res.RootId = sql.NullInt64{
			Valid: true,
			Int64: src.RootComment.Id,
		}
	}
	if src.ParentComment != nil {
		res.PID = sql.NullInt64{
			Valid: true,
			Int64: src.ParentComment.Id,
		}
	}
	res.Ctime = time.Now().UnixMilli()
	res.Utime = time.Now().UnixMilli()
	return res
}
