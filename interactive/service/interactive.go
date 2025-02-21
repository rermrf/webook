package service

import (
	"context"
	"golang.org/x/sync/errgroup"
	"webook/interactive/domain"
	"webook/interactive/repository"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=../../internal/service/mocks/interactive_mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	// Collect 收藏夹，cid 为收藏夹 id
	// cid 不一定有，或者是 0 对应的是该用户的默认收藏夹
	Collect(ctx context.Context, biz string, bizId int64, cid, uid int64) error
	CancelCollect(ctx context.Context, biz string, bizId int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		repo: repo,
	}
}

func (i *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	intrs, err := i.repo.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func (i *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

// Collect 收藏
func (i *interactiveService) Collect(ctx context.Context, biz string, bizId int64, cid, uid int64) error {
	// service 层面上是收藏
	return i.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (i *interactiveService) CancelCollect(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.RemoveCollectionItem(ctx, biz, bizId, uid)
}

func (i *interactiveService) Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error) {
	var intr domain.Interactive
	var liked bool
	var collected bool
	var eg errgroup.Group
	eg.Go(func() error {
		var err error
		intr, err = i.repo.Get(ctx, biz, bizId)
		return err
	})
	eg.Go(func() error {
		var err error
		liked, err = i.repo.Liked(ctx, biz, bizId, uid)
		return err
	})
	eg.Go(func() error {
		var err error
		collected, err = i.repo.Collected(ctx, biz, bizId, uid)
		return err
	})
	err := eg.Wait()
	if err != nil {
		return domain.Interactive{}, err
	}
	intr.Liked = liked
	intr.Collected = collected
	return intr, nil
}
