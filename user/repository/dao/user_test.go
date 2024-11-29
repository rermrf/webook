package dao

import (
	"context"
	"database/sql"
	"errors"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGormUserDao_Insert(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(t *testing.T) *sql.DB
		ctx     context.Context
		user    User
		wantErr error
		wantId  int64
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				res := sqlmock.NewResult(3, 1)
				// 这里是正则表达式，只要是匹配到以 INSERT INTO `users` 的语句
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
			ctx: context.Background(),
			user: User{
				Email: sql.NullString{
					String: "test@test.com",
					Valid:  true,
				},
				Password: "password",
				Phone: sql.NullString{
					String: "15011111111",
					Valid:  true,
				},
			},
			wantErr: nil,
		},
		{
			name: "邮箱冲突或者手机号冲突",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 这里是正则表达式，只要是匹配到以 INSERT INTO `users` 的语句
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(&mysql.MySQLError{Number: 1062})
				require.NoError(t, err)
				return mockDB
			},
			ctx:     context.Background(),
			user:    User{},
			wantErr: ErrUserDuplicate,
		},
		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 这里是正则表达式，只要是匹配到以 INSERT INTO `users` 的语句
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(errors.New("database error"))
				require.NoError(t, err)
				return mockDB
			},
			ctx:     context.Background(),
			user:    User{},
			wantErr: errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockDB := tc.mock(t)

			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn:                      mockDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				// mock DB 不需要 ping
				DisableAutomaticPing: true,
				// 跳过 gorm 的默认开启事物
				SkipDefaultTransaction: true,
			})
			require.NoError(t, err)
			ud := NewUserDao(db)
			err = ud.Insert(tc.ctx, tc.user)

			assert.Equal(t, tc.wantErr, err)
		})
	}
}
