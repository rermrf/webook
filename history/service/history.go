package service

import (
	"context"

	"webook/history/domain"
	"webook/history/repository"
)

type HistoryService interface {
	Record(ctx context.Context, h domain.BrowseHistory) error
	List(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.BrowseHistory, bool, error)
	Clear(ctx context.Context, userId int64) error
}

type historyService struct {
	repo repository.HistoryRepository
}

func NewHistoryService(repo repository.HistoryRepository) HistoryService {
	return &historyService{repo: repo}
}

func (s *historyService) Record(ctx context.Context, h domain.BrowseHistory) error {
	return s.repo.Record(ctx, h)
}

func (s *historyService) List(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.BrowseHistory, bool, error) {
	// 多取一条用于判断是否还有更多数据
	records, err := s.repo.List(ctx, userId, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	return records, hasMore, nil
}

func (s *historyService) Clear(ctx context.Context, userId int64) error {
	return s.repo.Clear(ctx, userId)
}
