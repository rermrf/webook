package service

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/article/domain"
	service2 "webook/article/service"
	"webook/internal/repository"
)

type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingService struct {
	artSvc    service2.ArticleService
	intrSvc   intrv1.InteractiveServiceClient
	repo      repository.RankingRepository
	batchSize int
	n         int
	// scoreFunc 不能返回负数
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc service2.ArticleService, intrSvc intrv1.InteractiveServiceClient, repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		repo:      repo,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			ms := time.Since(t)
			return float64(likeCnt-1) / math.Pow(float64(ms+2), 1.5)
		},
	}
}

func (svc *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	// 在这里存起来
	return svc.repo.ReplaceTopN(ctx, arts)
}

func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	// 只取七天内的数据
	now := time.Now()
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	topN := queue.NewPriorityQueue[Score](svc.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})
	for {
		// 那一批文章的数据
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})
		// 要去拿对应的点赞数据
		intrs, err := svc.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz:    "artile",
			BizIds: ids,
		})
		if err != nil {
			return nil, err
		}
		// 合并计算 score
		// 排序
		for _, art := range arts {
			intr := intrs.Intrs[art.Id]
			//intr, ok := intrs[art.Id]
			//if !ok {
			//	continue
			//}
			score := svc.scoreFunc(art.Utime, intr.LikeCnt)
			// 要考虑，这个 score 在不在前一百名
			// 拿到热度最低的
			// 小顶堆，最顶上的是最小的

			err = topN.Enqueue(Score{art, score})

			if errors.Is(err, queue.ErrOutOfCapacity) {
				// 这种写法要求 topN 已经满了
				val, _ := topN.Dequeue()
				if val.score < score {
					_ = topN.Enqueue(Score{art, score})
				} else {
					_ = topN.Enqueue(val)
				}
			}
		}

		// 一批已经处理完了，要不要进入下一批，怎么知道还有没有
		if len(arts) < svc.batchSize || now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			// 这一批都还没取够，当然可以肯定没有下一批了
			// 又或者已经取到了七天之前的数据了，说明可以中断了
			break
		}
		// 更新 offset
		offset += len(arts)
	}
	// 得出结果
	res := make([]domain.Article, svc.n)
	for i := svc.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			// 说明取完了，不够 n
			break
		}
		res[i] = val.art
	}
	return res, nil
}
