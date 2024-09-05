package repository

import (
	"context"
	"database/sql"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
}

type CachedUserRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDao, c cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (repo *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从 cache 查找
	// 再从 dao 里面找
	// 找到了回写到 cache
	user, err := repo.cache.Get(ctx, id)
	// - 缓存里有数据
	// - 缓存里没有数据
	// - 缓存出错
	if err == nil {
		return user, nil
	}
	// 没这个数据
	//if err == cache.ErrKeyNotExist {
	//	// 数据库查找
	//	// select * from users where id = ?
	//	u, err := r.dao.FindById(ctx, id)
	//	if err != nil {
	//		return domain.User{}, err
	//	}
	//	return domain.User{
	//		Id:    u.Id,
	//		Email: u.Email,
	//	}, nil
	//}

	u, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	user = repo.entityToDomain(u)

	go func() {
		err = repo.cache.Set(ctx, user)
		if err != nil {
			// 这里并不需要管，打日志，做监控就好
			//return domain.User{}, err
		}
	}()

	return user, err

	// 缓存出错，比如：err = io.EOF
	// 比如出现 缓存击穿、缓存雪崩，如果直接访问mysql，则数据库可能会崩

	// - 如果加载 -- 做好兜底，万一 redis 真的崩了，你要保护住你的数据库
	// -- 数据库限流
	// - 选不加载 -- 用户体验差一点
}

func (repo *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE email = ?
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	user := repo.entityToDomain(u)

	return user, nil
}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	// SELECT * FROM `users` WHERE phone = ?
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}

	user := repo.entityToDomain(u)

	return user, nil
}

func (repo *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(u))
	// 在这里操作缓存
}

func (repo *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
	}
}

func (repo *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
