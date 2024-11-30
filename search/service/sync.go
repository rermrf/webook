package service

import (
	"context"
	"webook/search/domain"
	"webook/search/repository"
)

type syncService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
	anyRepo     repository.AnyRepository
}

func NewSyncService(userRepo repository.UserRepository, articleRepo repository.ArticleRepository, anyRepo repository.AnyRepository) SyncService {
	return &syncService{userRepo: userRepo, articleRepo: articleRepo, anyRepo: anyRepo}
}

func (s *syncService) InputArticle(ctx context.Context, article domain.Article) error {
	return s.articleRepo.InputArticle(ctx, article)
}

func (s *syncService) InputUser(ctx context.Context, user domain.User) error {
	return s.userRepo.InputUser(ctx, user)
}

func (s *syncService) InputAny(ctx context.Context, index, docId, data string) error {
	return s.anyRepo.Input(ctx, index, docId, data)
}
