package repository

import (
	"context"

	"webook/notification/domain"
	"webook/notification/repository/dao"
)

// TransactionRepository 事务消息仓储接口
type TransactionRepository interface {
	Create(ctx context.Context, t domain.Transaction) (int64, error)
	FindByKey(ctx context.Context, key string) (domain.Transaction, error)
	UpdateStatus(ctx context.Context, key string, status domain.TransactionStatus) error
	FindPreparedTimeout(ctx context.Context, limit int) ([]domain.Transaction, error)
	IncrRetryCount(ctx context.Context, id int64, nextCheckTime int64) error
}

type TransactionRepositoryImpl struct {
	dao dao.TransactionDAO
}

func NewTransactionRepository(dao dao.TransactionDAO) TransactionRepository {
	return &TransactionRepositoryImpl{dao: dao}
}

func (r *TransactionRepositoryImpl) Create(ctx context.Context, t domain.Transaction) (int64, error) {
	return r.dao.Insert(ctx, r.toEntity(t))
}

func (r *TransactionRepositoryImpl) FindByKey(ctx context.Context, key string) (domain.Transaction, error) {
	entity, err := r.dao.FindByKey(ctx, key)
	if err != nil {
		return domain.Transaction{}, err
	}
	return r.toDomain(entity), nil
}

func (r *TransactionRepositoryImpl) UpdateStatus(ctx context.Context, key string, status domain.TransactionStatus) error {
	return r.dao.UpdateStatus(ctx, key, uint8(status))
}

func (r *TransactionRepositoryImpl) FindPreparedTimeout(ctx context.Context, limit int) ([]domain.Transaction, error) {
	entities, err := r.dao.FindPreparedTimeout(ctx, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Transaction, 0, len(entities))
	for _, e := range entities {
		result = append(result, r.toDomain(e))
	}
	return result, nil
}

func (r *TransactionRepositoryImpl) IncrRetryCount(ctx context.Context, id int64, nextCheckTime int64) error {
	return r.dao.IncrRetryCount(ctx, id, nextCheckTime)
}

func (r *TransactionRepositoryImpl) toEntity(t domain.Transaction) dao.NotificationTransaction {
	return dao.NotificationTransaction{
		Id:                 t.Id,
		NotificationId:     t.NotificationId,
		KeyField:           t.Key,
		BizId:              t.BizId,
		Status:             uint8(t.Status),
		CheckBackTimeoutMs: t.CheckBackTimeoutMs,
		NextCheckTime:      t.NextCheckTime,
		RetryCount:         t.RetryCount,
		MaxRetry:           t.MaxRetry,
	}
}

func (r *TransactionRepositoryImpl) toDomain(e dao.NotificationTransaction) domain.Transaction {
	return domain.Transaction{
		Id:                 e.Id,
		NotificationId:     e.NotificationId,
		Key:                e.KeyField,
		BizId:              e.BizId,
		Status:             domain.TransactionStatus(e.Status),
		CheckBackTimeoutMs: e.CheckBackTimeoutMs,
		NextCheckTime:      e.NextCheckTime,
		RetryCount:         e.RetryCount,
		MaxRetry:           e.MaxRetry,
		Ctime:              e.Ctime,
		Utime:              e.Utime,
	}
}
