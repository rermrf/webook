package integration

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"testing"
	"time"
	pmtv1 "webook/api/proto/gen/payment/v1"
	pmtmocks "webook/api/proto/gen/payment/v1/mocks"
	"webook/reward/domain"
	"webook/reward/integration/startup"
	"webook/reward/repository/dao"
)

type WechatNativeRewardServiceTestSuite struct {
	suite.Suite
	rdb redis.Cmdable
	db  *gorm.DB
}

func (s *WechatNativeRewardServiceTestSuite) SetupSuite() {
	s.rdb = startup.InitRedis()
	s.db = startup.InitDB()
}

func (s *WechatNativeRewardServiceTestSuite) TearDownSuite() {
	s.db.Exec("TRUNCATE TABLE `rewards`")
}

func (s *WechatNativeRewardServiceTestSuite) TestPreReward() {
	testCases := []struct {
		name   string
		mock   func(ctrl *gomock.Controller) pmtv1.WechatPaymentServiceClient
		before func(t *testing.T)
		after  func(t *testing.T)

		r domain.Reward

		wantData string
		wantErr  error
	}{
		{
			name: "直接创建成功",
			mock: func(ctrl *gomock.Controller) pmtv1.WechatPaymentServiceClient {
				client := pmtmocks.NewMockWechatPaymentServiceClient(ctrl)
				client.EXPECT().NativePrePay(gomock.Any(), gomock.Any()).
					Return(&pmtv1.NativePrePayResponse{
						CodeUrl: "test_url",
					}, nil)
				return client
			},
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				var r dao.Reward
				err := s.db.WithContext(ctx).Where("biz = ? and biz_id = ?", "test", 1).First(&r).Error
				assert.NoError(s.T(), err)
				assert.True(t, r.Id > 0)
				r.Id = 0
				assert.True(t, r.Ctime > 0)
				r.Ctime = 0
				assert.True(t, r.Utime > 0)
				r.Utime = 0
				assert.Equal(t, dao.Reward{
					Biz:       "test",
					BizId:     1,
					BizName:   "测试",
					TargetUid: 1234,
					Uid:       123,
					Amount:    1,
				}, r)
			},
			r: domain.Reward{
				Uid: 123,
				Target: domain.Target{
					Biz:     "test",
					BizId:   1,
					BizName: "测试",
					Uid:     1234,
				},
				Amt: 1,
			},
			wantData: "test_url",
		},
		{
			name: "拿到缓存",
			mock: func(ctrl *gomock.Controller) pmtv1.WechatPaymentServiceClient {
				client := pmtmocks.NewMockWechatPaymentServiceClient(ctrl)
				client.EXPECT().NativePrePay(gomock.Any(), gomock.Any()).
					Return(&pmtv1.NativePrePayResponse{
						CodeUrl: "test_url_1",
					}, nil)
				return client
			},
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				err := s.rdb.Set(ctx, s.codeURLKey("test", 2, 122), "test_url_1", time.Minute).Err()
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				codeUrl, err := s.rdb.GetDel(ctx, s.codeURLKey("test", 2, 122)).Result()
				assert.NoError(s.T(), err)
				assert.Equal(t, "test_url_1", codeUrl)
			},
			r: domain.Reward{
				Uid: 122,
				Target: domain.Target{
					Biz:     "test",
					BizId:   2,
					BizName: "测试",
					Uid:     1233,
				},
				Amt: 1,
			},
			wantData: "test_url_1",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()
			svc := startup.InitWechatNativeSvc(tc.mock(ctrl))
			codeURL, err := svc.PreReward(context.Background(), tc.r)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantData, codeURL.URL)
		})
	}
}

func (s *WechatNativeRewardServiceTestSuite) codeURLKey(biz string, bizId int, uId int) string {
	return fmt.Sprintf("reward:code_url:%s:%d:%d", biz, bizId, uId)
}

func TestWechatNativeRewardService(t *testing.T) {
	suite.Run(t, new(WechatNativeRewardServiceTestSuite))
}
