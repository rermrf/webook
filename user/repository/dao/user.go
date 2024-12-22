package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicate = errors.New("邮箱/手机已被注册")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

//go:generate mockgen -source=./user.go -package=daomocks -destination=mocks/user_dao_mock.go
type UserDao interface {
	FindById(ctx context.Context, id int64) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	Insert(ctx context.Context, u User) error
	UpdateNonZeroFields(ctx context.Context, u User) error
	FindByWechat(ctx context.Context, openID string) (User, error)
	FindByIds(ctx context.Context, ids []int64) ([]User, error)
}

type GormUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &GormUserDao{db: db}
}

func (dao *GormUserDao) FindByWechat(ctx context.Context, openID string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openID).First(&user).Error
	return user, err
}

func (dao *GormUserDao) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (dao *GormUserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *GormUserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

func (dao *GormUserDao) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		const uniqueConflictsErrNo = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突或者手机号冲突
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *GormUserDao) UpdateNonZeroFields(ctx context.Context, u User) error {
	return dao.db.WithContext(ctx).Updates(&u).Error
}

func (dao *GormUserDao) FindByIds(ctx context.Context, ids []int64) ([]User, error) {
	var users []User
	err := dao.db.WithContext(ctx).Where("id in (?)", ids).Find(&users).Error
	return users, err
}

// User 直接对应数据库表结构
type User struct {
	Id       int64          `gorm:"primaryKey;autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Phone    sql.NullString `gorm:"unique"`
	Password string

	// 索引的最左匹配原则：
	// 假如索引在 <A, B, C> 建好了
	// A, AB, ABC 都能用
	// WHERE 里面带了 ABC，可以用
	// WHERE 里面没有 A，就不能用

	// 如果创建联合索引，<unionid, openid> 用 openid 查询的时候不会走索引
	// <openid, unionid> 用 unionid查询的时候，不会走索引
	// 微信的字段
	WechatUnionID sql.NullString `gorm:"unique"`
	WechatOpenID  sql.NullString `gorm:"unique"`
	Nickname      string
	AboutMe       string
	Birthday      int64
	Ctime         int64
	Utime         int64
}
