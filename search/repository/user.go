package repository

import (
	"context"
	"webook/search/domain"
	"webook/search/repository/dao"
)

type userRepository struct {
	dao dao.UserDao
}

func NewUserRepository(dao dao.UserDao) UserRepository {
	return &userRepository{dao: dao}
}

func (u *userRepository) InputUser(ctx context.Context, msg domain.User) error {
	return u.dao.InputUser(ctx, dao.User{
		Id:       msg.Id,
		Email:    msg.Email,
		Nickname: msg.Nickname,
		Phone:    msg.Phone,
	})
}

func (u *userRepository) SearchUser(ctx context.Context, keywords []string) ([]domain.User, error) {
	users, err := u.dao.Search(ctx, keywords)
	if err != nil {
		return nil, err
	}
	res := make([]domain.User, 0, len(users))
	for _, user := range users {
		res = append(res, domain.User{
			Id:       user.Id,
			Email:    user.Email,
			Nickname: user.Nickname,
			Phone:    user.Phone,
		})
	}
	return res, nil
}
