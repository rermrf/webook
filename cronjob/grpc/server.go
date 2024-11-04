package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	cronjobv1 "webook/api/proto/gen/cronjob/v1"
	"webook/cronjob/domain"
	"webook/cronjob/service"
)

type CronJobServiceServer struct {
	svc service.JobService
	cronjobv1.UnimplementedCronJobServiceServer
}

func NewCronJobServiceServer(svc service.JobService) *CronJobServiceServer {
	return &CronJobServiceServer{svc: svc}
}

func (c *CronJobServiceServer) Register(server *grpc.Server) {
	cronjobv1.RegisterCronJobServiceServer(server, c)
}

func (c *CronJobServiceServer) Preempt(ctx context.Context, request *cronjobv1.PreemptRequest) (*cronjobv1.PreemptResponse, error) {
	job, err := c.svc.Preempt(ctx)
	if err != nil {
		return nil, err
	}
	return &cronjobv1.PreemptResponse{
		Job: c.toV(job),
	}, nil
}

func (c *CronJobServiceServer) ResetNextTime(ctx context.Context, request *cronjobv1.ResetNextTimeRequest) (*cronjobv1.ResetNextTimeResponse, error) {
	err := c.svc.ResetNextTime(ctx, c.toDomain(request.GetJob()))
	return &cronjobv1.ResetNextTimeResponse{}, err
}

func (c *CronJobServiceServer) AddJob(ctx context.Context, request *cronjobv1.AddJobRequest) (*cronjobv1.AddJobResponse, error) {
	err := c.svc.AddJob(ctx, c.toDomain(request.GetJob()))
	return &cronjobv1.AddJobResponse{}, err
}

func (c *CronJobServiceServer) toDomain(job *cronjobv1.CronJob) domain.Job {
	return domain.Job{
		Id:       job.GetId(),
		Name:     job.GetName(),
		Cron:     job.GetCron(),
		Executor: job.GetExecutor(),
		Cfg:      job.GetCfg(),
		NextTime: job.GetNextTime().AsTime(),
	}
}

func (c *CronJobServiceServer) toV(job domain.Job) *cronjobv1.CronJob {
	return &cronjobv1.CronJob{
		Id:       job.Id,
		Name:     job.Name,
		Cron:     job.Cron,
		Executor: job.Executor,
		Cfg:      job.Cfg,
		NextTime: timestamppb.New(job.NextTime),
	}
}
