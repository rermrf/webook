package service

import (
	"context"
	"webook/account/domain"
	"webook/account/repository"
)

type accountService struct {
	repo repository.AccountRepository
}

func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

func (a *accountService) Credit(ctx context.Context, cr domain.Credit) error {
	// 需要做幂等：唯一索引冲突（兜底）、布隆过滤器、redis 里面看一下有没有这个 biz + biz_id
	return a.repo.AddCredit(ctx, cr)
}
