package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDao
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDao, c *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: c,
	}
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从 cache 查找
	// 再从 dao 里面找
	// 找到了回写到 cache
	user, err := r.cache.Get(ctx, id)
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

	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	user = domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}

	go func() {
		err = r.cache.Set(ctx, user)
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

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE email = ?
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
	// 在这里操作缓存
}
