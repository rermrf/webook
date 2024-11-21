package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
	"webook/account/grpc"
	"webook/account/integration/startup"
	"webook/account/repository/dao"
	accountv1 "webook/api/proto/gen/account/v1"
)

type AccountServiceServerTestSuite struct {
	suite.Suite
	db     *gorm.DB
	server *grpc.AccountServiceServer
}

func (s *AccountServiceServerTestSuite) SetupSuite() {
	s.db = startup.InitDB()
	s.server = startup.InitAccountServiceServer()
}

func (s *AccountServiceServerTestSuite) TearDownSuite() {
	s.db.Exec("TRUNCATE TABLE `accounts`")
}

func (s *AccountServiceServerTestSuite) TestCredit() {
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		req     *accountv1.CreditRequest
		wantErr error
	}{
		{
			name: "用户账号不存在",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
				defer cancel()
				var sysAccount dao.Account
				err := s.db.WithContext(ctx).Where("type = ?", uint8(accountv1.AccountType_AccountTypeSystem)).First(&sysAccount).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(10), sysAccount.Balance)

				var userAccount dao.Account
				err = s.db.WithContext(ctx).Where("uid = ?", 1022).First(&userAccount).Error
				assert.NoError(t, err)
				userAccount.Id = 0
				assert.True(t, userAccount.Ctime > 0)
				userAccount.Ctime = 0
				assert.True(t, userAccount.Utime > 0)
				userAccount.Utime = 0
				assert.Equal(t, dao.Account{
					Account:  770,
					Uid:      1022,
					Type:     uint8(accountv1.AccountType_AccountTypeReward),
					Balance:  100,
					Currency: "CNY",
				}, userAccount)
			},
			req: &accountv1.CreditRequest{
				Biz:   "test",
				BizId: 123,
				Items: []*accountv1.CreditItem{
					{
						Account:     770,
						Uid:         1022,
						AccountType: accountv1.AccountType_AccountTypeReward,
						// 金额数
						Amt:      100,
						Currency: "CNY",
					},
					{
						AccountType: accountv1.AccountType_AccountTypeSystem,
						Amt:         10,
						Currency:    "CNY",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "用户账号存在",
			before: func(t *testing.T) {
				now := time.Now().UnixMilli()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(&dao.Account{
					Uid:      577,
					Account:  123,
					Type:     uint8(accountv1.AccountType_AccountTypeReward),
					Balance:  300,
					Currency: "CNY",
					Ctime:    now,
					Utime:    now,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var sysAccount dao.Account
				err := s.db.WithContext(ctx).Where("type = ?", uint8(accountv1.AccountType_AccountTypeSystem)).First(&sysAccount).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(100), sysAccount.Balance)

				var userAccount dao.Account
				err = s.db.WithContext(ctx).Where("uid = ?", 577).First(&userAccount).Error
				assert.NoError(t, err)
				userAccount.Id = 0
				assert.True(t, userAccount.Ctime > 0)
				userAccount.Ctime = 0
				assert.True(t, userAccount.Utime > 0)
				userAccount.Utime = 0
				assert.Equal(t, dao.Account{
					Account:  123,
					Uid:      577,
					Type:     uint8(accountv1.AccountType_AccountTypeReward),
					Balance:  1300,
					Currency: "CNY",
				}, userAccount)
			},
			req: &accountv1.CreditRequest{
				Biz:   "test",
				BizId: 1234,
				Items: []*accountv1.CreditItem{
					{
						Account:     123,
						Uid:         577,
						AccountType: accountv1.AccountType_AccountTypeReward,
						// 金额数
						Amt:      1000,
						Currency: "CNY",
					},
					{
						AccountType: accountv1.AccountType_AccountTypeSystem,
						Amt:         100,
						Currency:    "CNY",
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			_, err := s.server.Credit(context.Background(), tc.req)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func TestAccountServiceServer(t *testing.T) {
	suite.Run(t, new(AccountServiceServerTestSuite))
}
