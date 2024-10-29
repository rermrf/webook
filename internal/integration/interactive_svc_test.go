package integration

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
	"webook/internal/integration/startup"
	"webook/internal/repository/dao"
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

func TestInteractiveSvcTestSuite(t *testing.T) {
	suite.Run(t, &InteractiveSvcTestSuite{})
}
