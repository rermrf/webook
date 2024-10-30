package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	domain2 "webook/interactive/domain"
	service2 "webook/interactive/service"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
	svcmocks "webook/internal/service/mocks"
)

func TestRankingTopN(t *testing.T) {
	now := time.Now()
	testcases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (ArticleService, service2.InteractiveService, repository.RankingRepository)
		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "计算成功",
			mock: func(ctrl *gomock.Controller) (ArticleService, service2.InteractiveService, repository.RankingRepository) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				// 最简单的，一批搞完
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 3).Return([]domain.Article{
					domain.Article{
						Id:    1,
						Utime: now,
						Ctime: now,
					},
					domain.Article{
						Id:    2,
						Utime: now,
						Ctime: now,
					},
					domain.Article{
						Id:    3,
						Utime: now,
						Ctime: now,
					},
				}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 3, 3).Return([]domain.Article{}, nil)
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2, 3}).Return(map[int64]domain2.Interactive{
					1: {
						BizId:   1,
						LikeCnt: 1,
					},
					2: {
						BizId:   2,
						LikeCnt: 2,
					},
					3: {
						BizId:   3,
						LikeCnt: 3,
					},
				}, nil)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{}).Return(map[int64]domain2.Interactive{}, nil)
				rankingRepo := repomocks.NewMockRankingRepository(ctrl)
				return artSvc, intrSvc, rankingRepo
			},
			wantArts: []domain.Article{
				domain.Article{
					Id:    3,
					Utime: now,
					Ctime: now,
				},
				domain.Article{
					Id:    2,
					Utime: now,
					Ctime: now,
				},
				domain.Article{
					Id:    1,
					Utime: now,
					Ctime: now,
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			artSvc, intrSvc, repo := tc.mock(ctrl)
			rankSvc := NewBatchRankingService(artSvc, intrSvc, repo).(*BatchRankingService)
			// 为了测试
			rankSvc.batchSize = 3
			rankSvc.n = 3
			rankSvc.scoreFunc = func(t time.Time, likeCnt int64) float64 {
				return float64(likeCnt)
			}
			arts, err := rankSvc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
