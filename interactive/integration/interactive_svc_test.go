package integration

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
	"webook/interactive/domain"
	"webook/interactive/integration/startup"
	"webook/interactive/repository/dao"
)

type InteractiveSvcTestSuite struct {
	suite.Suite
	db  *gorm.DB
	rdb redis.Cmdable
}

func (s *InteractiveSvcTestSuite) SetupSuite() {
	s.db = startup.InitDB()
	s.rdb = startup.InitRedis()
}

func (s *InteractiveSvcTestSuite) TearDownSuite() {
	err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_like_bizs`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_collection_bizs`").Error
	assert.NoError(s.T(), err)
}

func (s *InteractiveSvcTestSuite) TestIncrReadCnt() {
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64

		wantErr error
	}{
		{
			name: "增加成功，db和redis中有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				// 数据库插入
				err := s.db.Create(dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
					Utime:      7,
				}).Error
				assert.NoError(s.T(), err)
				// redis插入 HSET KEY_NAME FIELD VALUE
				err = s.rdb.HSet(ctx, "interactive:test:2", "read_cnt", 3).Err()
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(t, data.Utime > 7)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:    1,
					Biz:   "test",
					BizId: 2,
					// +1
					ReadCnt:    4,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
				}, data)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "read_cnt").Int()
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), 4, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(s.T(), err)
			},
			biz:   "test",
			bizId: 2,
		},
		{
			name: "增加成功，db中有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				// 数据库插入
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         2,
					Biz:        "test",
					BizId:      3,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
					Utime:      7,
				}).Error
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 2).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(t, data.Utime > 7)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:    2,
					Biz:   "test",
					BizId: 3,
					// +1
					ReadCnt:    4,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:3").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
				err = s.rdb.Del(ctx, "interactive:test:3").Err()
				assert.NoError(s.T(), err)

			},
			biz:   "test",
			bizId: 3,
		},
		{
			name: "增加成功，都没有数据",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("biz = ? AND biz_id = ?", "test", 4).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(t, data.Id > 0)
				data.Id = 0
				assert.True(t, data.Utime > 0)
				data.Utime = 0
				assert.True(t, data.Ctime > 0)
				data.Ctime = 0
				assert.Equal(t, dao.Interactive{
					Biz:   "test",
					BizId: 4,
					// +1
					ReadCnt: 1,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:4").Result()
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), int64(0), cnt)
			},
			biz:   "test",
			bizId: 4,
		},
	}

	// 交互服务没有 http 请求，所以不需要每次循环都初始化
	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.IncrReadCnt(context.Background(), tc.biz, tc.bizId)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveSvcTestSuite) TestLike() {
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		uid   int64

		wantErr error
	}{
		{
			name: "点赞成功，数据库和缓存中都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         10,
					Biz:        "test",
					BizId:      20,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
					Utime:      7,
				}).Error
				assert.NoError(s.T(), err)
				err = s.rdb.HSet(ctx, "interactive:test:20", "like_cnt", 5).Err()
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				// 验证交互表中的点赞数量是否加一
				var data dao.Interactive
				err := s.db.Where("id = ?", 10).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), data.Utime > 7)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         10,
					Biz:        "test",
					BizId:      20,
					ReadCnt:    3,
					CollectCnt: 4,
					// +1
					LikeCnt: 6,
					Ctime:   6,
				}, data)
				// 验证用户点赞表是否存在记录
				var ul dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?", "test", 20, 10).First(&ul).Error
				assert.True(t, ul.Id > 0)
				ul.Id = 0
				assert.True(t, ul.Utime > 0)
				ul.Utime = 0
				assert.True(t, ul.Ctime > 0)
				ul.Ctime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Biz:    "test",
					BizId:  20,
					Uid:    10,
					Status: 1,
				}, ul)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:20", "like_cnt").Int64()
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), int64(6), cnt)
				err = s.rdb.Del(ctx, "interactive:test:20").Err()
				assert.NoError(s.T(), err)
			},
			biz:   "test",
			bizId: 20,
			uid:   10,
		},
		{
			name: "点赞成功，只有数据库有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         20,
					Biz:        "test",
					BizId:      30,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
					Utime:      7,
				}).Error
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				// 验证交互表中的点赞数量是否加一
				var data dao.Interactive
				err := s.db.WithContext(ctx).Where("id = ?", 20).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), data.Utime > 7)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         20,
					Biz:        "test",
					BizId:      30,
					ReadCnt:    3,
					CollectCnt: 4,
					// +1
					LikeCnt: 6,
					Ctime:   6,
				}, data)
				// 验证用户点赞表是否存在记录
				var ul dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?", "test", 30, 10).First(&ul).Error
				assert.True(t, ul.Id > 0)
				ul.Id = 0
				assert.True(t, ul.Utime > 0)
				ul.Utime = 0
				assert.True(t, ul.Ctime > 0)
				ul.Ctime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Biz:    "test",
					BizId:  30,
					Uid:    10,
					Status: 1,
				}, ul)

				cnt, err := s.rdb.Exists(ctx, "interactive:test:30").Result()
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), int64(0), cnt)
			},
			biz:   "test",
			bizId: 30,
			uid:   10,
		},
		{
			name: "点赞成功，数据库和缓存中都没有数据",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				// 验证交互表中的点赞数量是否加一
				var data dao.Interactive
				err := s.db.WithContext(ctx).Where("biz = ? AND biz_id = ?", "test", 40).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(t, data.Id > 0)
				data.Id = 0
				assert.True(t, data.Utime > 0)
				data.Utime = 0
				assert.True(t, data.Ctime > 0)
				data.Ctime = 0
				assert.Equal(t, dao.Interactive{
					Biz:   "test",
					BizId: 40,
					// +1
					LikeCnt: 1,
				}, data)
				// 验证用户点赞表是否存在记录
				var ul dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?", "test", 40, 10).First(&ul).Error
				assert.True(t, ul.Id > 0)
				ul.Id = 0
				assert.True(t, ul.Utime > 0)
				ul.Utime = 0
				assert.True(t, ul.Ctime > 0)
				ul.Ctime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Biz:    "test",
					BizId:  40,
					Uid:    10,
					Status: 1,
				}, ul)

				cnt, err := s.rdb.Exists(ctx, "interactive:test:40").Result()
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), int64(0), cnt)
			},
			biz:   "test",
			bizId: 40,
			uid:   10,
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.Like(context.Background(), tc.biz, tc.bizId, tc.uid)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveSvcTestSuite) TestCancelLike() {
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		biz     string
		bizId   int64
		uid     int64
		wantErr error
	}{
		{
			name: "取消点赞成功，db和cache中都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:      100,
					Biz:     "test",
					BizId:   200,
					LikeCnt: 300,
					Ctime:   8,
					Utime:   9,
				}).Error
				assert.NoError(s.T(), err)
				err = s.db.WithContext(ctx).Create(dao.UserLikeBiz{
					Id:     10,
					Biz:    "test",
					BizId:  200,
					Uid:    100,
					Ctime:  8,
					Utime:  9,
					Status: 1,
				}).Error
				assert.NoError(s.T(), err)
				err = s.rdb.HSet(ctx, "interactive:test:200", "like_cnt", 300).Err()
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var data dao.Interactive
				err := s.db.WithContext(ctx).Where("id = ?", 100).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), data.Utime > 0)
				data.Utime = 0

				assert.Equal(t, dao.Interactive{
					Id:      100,
					Biz:     "test",
					BizId:   200,
					LikeCnt: 299,
					Ctime:   8,
				}, data)

				var ul dao.UserLikeBiz
				err = s.db.Where("id = ?", 100).First(&ul).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), ul.Utime > 9)
				ul.Utime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Id:     10,
					Biz:    "test",
					BizId:  200,
					Uid:    100,
					Ctime:  8,
					Utime:  0,
					Status: 0,
				}, ul)
				err = s.rdb.HDel(ctx, "interactive:test:200", "like_cnt").Err()
				assert.NoError(s.T(), err)
			},
			biz:   "test",
			bizId: 200,
			uid:   100,
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := svc.CancelLike(ctx, tc.biz, tc.bizId, tc.uid)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func (s *InteractiveSvcTestSuite) TestCollect() {
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		biz     string
		bizId   int64
		cid     int64
		uid     int64
		wantErr error
	}{
		{
			name: "收藏成功，第一次收藏，db和cache都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         1001,
					Biz:        "test",
					BizId:      1002,
					ReadCnt:    1003,
					CollectCnt: 1004,
					LikeCnt:    1005,
					Ctime:      8,
					Utime:      9,
				}).Error
				assert.NoError(s.T(), err)
				err = s.rdb.HSet(ctx, "interactive:test:1002", "collect_cnt", 1004).Err()
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var data dao.Interactive
				err := s.db.WithContext(ctx).Where("id = ?", 1001).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), data.Utime > 9)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:      1001,
					Biz:     "test",
					BizId:   1002,
					ReadCnt: 1003,
					// +1
					CollectCnt: 1005,
					LikeCnt:    1005,
					Ctime:      8,
					Utime:      0,
				}, data)
				var uc dao.UserCollectionBiz
				err = s.db.WithContext(ctx).Where("biz = ? AND biz_id = ?", "test", 1002).First(&uc).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), uc.Utime > 0)
				uc.Utime = 0
				assert.True(t, uc.Ctime > 0)
				uc.Ctime = 0
				assert.True(t, uc.Id > 0)
				uc.Id = 0
				assert.Equal(t, dao.UserCollectionBiz{
					Biz:    "test",
					BizId:  1002,
					Cid:    111,
					Uid:    100,
					Status: 1,
				}, uc)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:1002", "collect_cnt").Int()
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), 1005, cnt)
				err = s.rdb.HDel(ctx, "interactive:test:1002", "collect_cnt").Err()
				assert.NoError(s.T(), err)
			},
			biz:   "test",
			bizId: 1002,
			uid:   100,
			cid:   111,
		},
		{
			name: "收藏成功，再次收藏，db和cache都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         1011,
					Biz:        "test",
					BizId:      1012,
					ReadCnt:    1003,
					CollectCnt: 1004,
					LikeCnt:    1005,
					Ctime:      8,
					Utime:      9,
				}).Error
				assert.NoError(s.T(), err)
				// 收藏表预置数据
				err = s.db.Create(dao.UserCollectionBiz{
					Id:     1000,
					Biz:    "test",
					BizId:  1012,
					Uid:    100,
					Cid:    111,
					Status: 0,
					Ctime:  8,
					Utime:  9,
				}).Error
				err = s.rdb.HSet(ctx, "interactive:test:1012", "collect_cnt", 1004).Err()
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var data dao.Interactive
				err := s.db.WithContext(ctx).Where("id = ?", 1011).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), data.Utime > 9)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:      1011,
					Biz:     "test",
					BizId:   1012,
					ReadCnt: 1003,
					// +1
					CollectCnt: 1005,
					LikeCnt:    1005,
					Ctime:      8,
					Utime:      0,
				}, data)
				var uc dao.UserCollectionBiz
				err = s.db.WithContext(ctx).Where("id = ?", 1000).First(&uc).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), uc.Utime > 0)
				uc.Utime = 0
				assert.True(t, uc.Ctime > 0)
				uc.Ctime = 0
				assert.True(t, uc.Id > 0)
				uc.Id = 0
				assert.Equal(t, dao.UserCollectionBiz{
					Biz:    "test",
					BizId:  1012,
					Cid:    111,
					Uid:    100,
					Status: 1,
				}, uc)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:1012", "collect_cnt").Int()
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), 1005, cnt)
				err = s.rdb.HDel(ctx, "interactive:test:1012", "collect_cnt").Err()
				assert.NoError(s.T(), err)
			},
			biz:   "test",
			bizId: 1012,
			uid:   100,
			cid:   111,
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.Collect(context.Background(), tc.biz, tc.bizId, tc.cid, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveSvcTestSuite) TestCancelCollect() {
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		biz     string
		bizId   int64
		uid     int64
		wantErr error
	}{
		{
			name: "取消收藏成功，db和cache都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         10000,
					Biz:        "test",
					BizId:      10001,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      8,
					Utime:      9,
				}).Error
				assert.NoError(s.T(), err)
				// 收藏表预置数据
				err = s.db.Create(dao.UserCollectionBiz{
					Id:     10001,
					Biz:    "test",
					BizId:  10001,
					Uid:    100,
					Cid:    1111,
					Status: 1,
					Ctime:  8,
					Utime:  9,
				}).Error
				err = s.rdb.HSet(ctx, "interactive:test:10001", "collect_cnt", 4).Err()
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var data dao.Interactive
				err := s.db.WithContext(ctx).Where("id = ?", 10000).First(&data).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), data.Utime > 9)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:      10000,
					Biz:     "test",
					BizId:   10001,
					ReadCnt: 3,
					// -1
					CollectCnt: 3,
					LikeCnt:    5,
					Ctime:      8,
					Utime:      0,
				}, data)
				var uc dao.UserCollectionBiz
				err = s.db.WithContext(ctx).Where("id = ?", 10001).First(&uc).Error
				assert.NoError(s.T(), err)
				assert.True(s.T(), uc.Utime > 0)
				uc.Utime = 0
				assert.True(t, uc.Ctime > 0)
				uc.Ctime = 0
				assert.True(t, uc.Id > 0)
				uc.Id = 0
				assert.Equal(t, dao.UserCollectionBiz{
					Biz:    "test",
					BizId:  10001,
					Cid:    1111,
					Uid:    100,
					Status: 0,
				}, uc)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:10001", "collect_cnt").Int()
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), 3, cnt)
				err = s.rdb.HDel(ctx, "interactive:test:10001", "collect_cnt").Err()
				assert.NoError(s.T(), err)
			},
			biz:   "test",
			bizId: 10001,
			uid:   100,
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.CancelCollect(context.Background(), tc.biz, tc.bizId, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveSvcTestSuite) TestGet() {
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		biz     string
		bizId   int64
		uid     int64
		wantRes domain.Interactive
		wantErr error
	}{
		{
			name: "获取到交互信息，已点赞，已收藏，db和cache中都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         12101,
					Biz:        "test",
					BizId:      1101,
					ReadCnt:    1000,
					LikeCnt:    1001,
					CollectCnt: 1002,
					Ctime:      8,
					Utime:      9,
				}).Error
				assert.NoError(s.T(), err)
				err = s.db.WithContext(ctx).Create(dao.UserLikeBiz{
					Id:     123,
					Biz:    "test",
					BizId:  1101,
					Uid:    577,
					Status: 1,
					Ctime:  8,
					Utime:  9,
				}).Error
				assert.NoError(s.T(), err)
				err = s.db.WithContext(ctx).Create(dao.UserCollectionBiz{
					Id:     123,
					Biz:    "test",
					BizId:  1101,
					Uid:    577,
					Cid:    0, // 默认收藏夹
					Status: 1,
					Ctime:  8,
					Utime:  9,
				}).Error
				assert.NoError(s.T(), err)
				err = s.rdb.HSet(ctx, "interactive:test:1101", "read_cnt", 1000).Err()
				assert.NoError(s.T(), err)
				err = s.rdb.HSet(ctx, "interactive:test:1101", "like_cnt", 1001).Err()
				assert.NoError(s.T(), err)
				err = s.rdb.HSet(ctx, "interactive:test:1101", "collect_cnt", 1002).Err()
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				s.rdb.Del(ctx, "interactive:test:1101")
			},
			biz:   "test",
			bizId: 1101,
			uid:   577,
			wantRes: domain.Interactive{
				Biz:        "test",
				BizId:      1101,
				ReadCnt:    1000,
				LikeCnt:    1001,
				CollectCnt: 1002,
				Liked:      true,
				Collected:  true,
			},
			wantErr: nil,
		},
		{
			name: "获取到交互信息，已点赞，已收藏，只有db中都有数据",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         13101,
					Biz:        "test",
					BizId:      1102,
					ReadCnt:    1000,
					LikeCnt:    1001,
					CollectCnt: 1002,
					Ctime:      8,
					Utime:      9,
				}).Error
				assert.NoError(s.T(), err)
				err = s.db.WithContext(ctx).Create(dao.UserLikeBiz{
					Id:     1234,
					Biz:    "test",
					BizId:  1102,
					Uid:    577,
					Status: 1,
					Ctime:  8,
					Utime:  9,
				}).Error
				assert.NoError(s.T(), err)
				err = s.db.WithContext(ctx).Create(dao.UserCollectionBiz{
					Id:     1234,
					Biz:    "test",
					BizId:  1102,
					Uid:    577,
					Cid:    0, // 默认收藏夹
					Status: 1,
					Ctime:  8,
					Utime:  9,
				}).Error
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				s.rdb.Del(ctx, "interactive:test:1102")
			},
			biz:   "test",
			bizId: 1102,
			uid:   577,
			wantRes: domain.Interactive{
				Biz:        "test",
				BizId:      1102,
				ReadCnt:    1000,
				LikeCnt:    1001,
				CollectCnt: 1002,
				Liked:      true,
				Collected:  true,
			},
			wantErr: nil,
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			res, err := svc.Get(context.Background(), tc.biz, tc.bizId, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, res)
			tc.after(t)
		})
	}
}

func TestInteractiveSvcTestSuite(t *testing.T) {
	suite.Run(t, &InteractiveSvcTestSuite{})
}
