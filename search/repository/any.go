package repository

import (
	"context"
	"webook/search/repository/dao"
)

type anyRepository struct {
	dao dao.AnyDao
}

func NewAnyRepository(dao dao.AnyDao) AnyRepository {
	return &anyRepository{dao: dao}
}

func (a *anyRepository) Input(ctx context.Context, index string, docId string, data string) error {
	return a.dao.Input(ctx, index, docId, data)
}
