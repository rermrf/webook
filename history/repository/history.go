package repository

import (
	"context"

	"webook/history/domain"
	"webook/history/repository/dao"
)

type HistoryRepository interface {
	Record(ctx context.Context, h domain.BrowseHistory) error
	List(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.BrowseHistory, error)
	Clear(ctx context.Context, userId int64) error
}

type historyRepository struct {
	dao dao.HistoryDAO
}

func NewHistoryRepository(dao dao.HistoryDAO) HistoryRepository {
	return &historyRepository{dao: dao}
}

func (r *historyRepository) Record(ctx context.Context, h domain.BrowseHistory) error {
	return r.dao.Upsert(ctx, r.toEntity(h))
}

func (r *historyRepository) List(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.BrowseHistory, error) {
	records, err := r.dao.FindByUserId(ctx, userId, cursor, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.BrowseHistory, 0, len(records))
	for _, record := range records {
		res = append(res, r.toDomain(record))
	}
	return res, nil
}

func (r *historyRepository) Clear(ctx context.Context, userId int64) error {
	return r.dao.DeleteByUserId(ctx, userId)
}

func (r *historyRepository) toEntity(h domain.BrowseHistory) dao.BrowseHistory {
	return dao.BrowseHistory{
		Id:         h.Id,
		UserId:     h.UserId,
		Biz:        h.Biz,
		BizId:      h.BizId,
		BizTitle:   h.BizTitle,
		AuthorName: h.AuthorName,
		Ctime:      h.Ctime,
		Utime:      h.Utime,
	}
}

func (r *historyRepository) toDomain(h dao.BrowseHistory) domain.BrowseHistory {
	return domain.BrowseHistory{
		Id:         h.Id,
		UserId:     h.UserId,
		Biz:        h.Biz,
		BizId:      h.BizId,
		BizTitle:   h.BizTitle,
		AuthorName: h.AuthorName,
		Ctime:      h.Ctime,
		Utime:      h.Utime,
	}
}
