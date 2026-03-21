package scheduler

import (
	"context"

	"webook/notification/domain"
	"webook/notification/repository"
	"webook/notification/service/channel"
	"webook/pkg/logger"
)

// ScheduledSendJob 延迟发送定时任务
// 实现 cronjobx.Job 接口，扫描到达发送时间的延迟通知并发送
type ScheduledSendJob struct {
	repo    repository.NotificationRepository
	senders map[domain.Channel]channel.Sender
	l       logger.LoggerV1
}

func NewScheduledSendJob(
	repo repository.NotificationRepository,
	senders map[domain.Channel]channel.Sender,
	l logger.LoggerV1,
) *ScheduledSendJob {
	return &ScheduledSendJob{
		repo:    repo,
		senders: senders,
		l:       l,
	}
}

func (j *ScheduledSendJob) Name() string {
	return "notification_scheduled_send"
}

func (j *ScheduledSendJob) Run(ctx context.Context) error {
	// 每次最多扫描 100 条到期的延迟通知
	ns, err := j.repo.FindScheduledReady(ctx, 100)
	if err != nil {
		j.l.Error("扫描延迟发送通知失败", logger.Error(err))
		return err
	}
	for _, n := range ns {
		j.sendOne(ctx, n)
	}
	return nil
}

func (j *ScheduledSendJob) sendOne(ctx context.Context, n domain.Notification) {
	// 更新状态为发送中
	err := j.repo.UpdateStatus(ctx, n.Id, domain.NotificationStatusSending)
	if err != nil {
		j.l.Error("更新通知状态失败",
			logger.Int64("id", n.Id),
			logger.Error(err))
		return
	}

	sender, ok := j.senders[n.Channel]
	if !ok {
		j.l.Error("不支持的渠道",
			logger.Int64("id", n.Id),
			logger.String("channel", n.Channel.String()))
		_ = j.repo.UpdateStatus(ctx, n.Id, domain.NotificationStatusFailed)
		return
	}

	err = sender.Send(ctx, n)
	if err != nil {
		j.l.Error("延迟发送通知失败",
			logger.Int64("id", n.Id),
			logger.Error(err))
		_ = j.repo.UpdateStatus(ctx, n.Id, domain.NotificationStatusFailed)
		return
	}

	_ = j.repo.UpdateStatus(ctx, n.Id, domain.NotificationStatusSent)
}
