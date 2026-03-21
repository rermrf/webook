package scheduler

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	notificationv2 "webook/api/proto/gen/notification/v2"
	"webook/notification/domain"
	"webook/notification/repository"
	"webook/notification/service"
	"webook/pkg/logger"
)

// CheckBackScheduler 事务回查调度器
// 实现 cronjobx.Job 接口，由外部 cron 调度器驱动执行
type CheckBackScheduler struct {
	txRepo        repository.TransactionRepository
	svc           service.NotificationService
	etcdClient    *clientv3.Client
	l             logger.LoggerV1
	maxRetry      int
	retryInterval time.Duration
}

func NewCheckBackScheduler(
	txRepo repository.TransactionRepository,
	svc service.NotificationService,
	etcdClient *clientv3.Client,
	l logger.LoggerV1,
) *CheckBackScheduler {
	return &CheckBackScheduler{
		txRepo:        txRepo,
		svc:           svc,
		etcdClient:    etcdClient,
		l:             l,
		maxRetry:      5,
		retryInterval: 10 * time.Second,
	}
}

// Name 实现 cronjobx.Job 接口
func (s *CheckBackScheduler) Name() string {
	return "notification_checkback"
}

// Run 实现 cronjobx.Job 接口，执行一次回查扫描
func (s *CheckBackScheduler) Run(ctx context.Context) error {
	txs, err := s.txRepo.FindPreparedTimeout(ctx, 100)
	if err != nil {
		s.l.Error("扫描超时事务失败", logger.Error(err))
		return err
	}
	for _, tx := range txs {
		s.checkOne(ctx, tx)
	}
	return nil
}

func (s *CheckBackScheduler) checkOne(ctx context.Context, tx domain.Transaction) {
	// 超过最大重试次数，强制取消
	if tx.RetryCount >= s.maxRetry {
		s.l.Warn("事务回查超过最大重试次数，强制取消",
			logger.String("key", tx.Key),
			logger.String("biz_id", tx.BizId),
		)
		err := s.svc.Cancel(ctx, tx.Key)
		if err != nil {
			s.l.Error("强制取消事务失败",
				logger.String("key", tx.Key),
				logger.Error(err),
			)
		}
		return
	}

	// 从 ETCD 中查找事务回查服务实例
	prefix := fmt.Sprintf("/services/transaction-checker/%s/", tx.BizId)
	resp, err := s.etcdClient.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithLimit(1))
	if err != nil {
		s.l.Error("从 ETCD 获取事务回查服务失败",
			logger.String("key", tx.Key),
			logger.String("biz_id", tx.BizId),
			logger.Error(err),
		)
		s.incrRetryCount(ctx, tx)
		return
	}
	if len(resp.Kvs) == 0 {
		s.l.Warn("未找到事务回查服务实例",
			logger.String("key", tx.Key),
			logger.String("biz_id", tx.BizId),
		)
		s.incrRetryCount(ctx, tx)
		return
	}

	// 获取服务地址
	addr := string(resp.Kvs[0].Value)

	// 建立 gRPC 连接
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		s.l.Error("连接事务回查服务失败",
			logger.String("key", tx.Key),
			logger.String("biz_id", tx.BizId),
			logger.String("addr", addr),
			logger.Error(err),
		)
		s.incrRetryCount(ctx, tx)
		return
	}
	defer conn.Close()

	// 调用 CheckTransaction
	client := notificationv2.NewTransactionCheckerClient(conn)
	checkResp, err := client.CheckTransaction(ctx, &notificationv2.CheckTransactionRequest{
		Key: tx.Key,
	})
	if err != nil {
		s.l.Error("调用事务回查失败",
			logger.String("key", tx.Key),
			logger.String("biz_id", tx.BizId),
			logger.Error(err),
		)
		s.incrRetryCount(ctx, tx)
		return
	}

	// 根据回查结果处理
	switch checkResp.GetAction() {
	case notificationv2.TransactionAction_TRANSACTION_ACTION_COMMIT:
		err = s.svc.Confirm(ctx, tx.Key)
		if err != nil {
			s.l.Error("确认事务失败",
				logger.String("key", tx.Key),
				logger.Error(err),
			)
		}
	case notificationv2.TransactionAction_TRANSACTION_ACTION_ROLLBACK:
		err = s.svc.Cancel(ctx, tx.Key)
		if err != nil {
			s.l.Error("取消事务失败",
				logger.String("key", tx.Key),
				logger.Error(err),
			)
		}
	case notificationv2.TransactionAction_TRANSACTION_ACTION_PENDING:
		s.incrRetryCount(ctx, tx)
	default:
		s.l.Warn("未知的事务回查结果",
			logger.String("key", tx.Key),
			logger.String("biz_id", tx.BizId),
		)
		s.incrRetryCount(ctx, tx)
	}
}

func (s *CheckBackScheduler) incrRetryCount(ctx context.Context, tx domain.Transaction) {
	nextCheckTime := time.Now().Add(s.retryInterval * time.Duration(tx.RetryCount+1)).UnixMilli()
	err := s.txRepo.IncrRetryCount(ctx, tx.Id, nextCheckTime)
	if err != nil {
		s.l.Error("增加重试次数失败",
			logger.String("key", tx.Key),
			logger.Error(err),
		)
	}
}
