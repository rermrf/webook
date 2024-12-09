package grpc

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"time"
	feedv1 "webook/api/proto/gen/feed"
	"webook/feed/domain"
	"webook/feed/service"
)

type FeedEventGrpcServer struct {
	feedv1.UnimplementedFeedSvcServer
	svc service.FeedService
}

func NewFeedEventGrpcServer(svc service.FeedService) *FeedEventGrpcServer {
	return &FeedEventGrpcServer{svc: svc}
}

func (f *FeedEventGrpcServer) Register(server grpc.ServiceRegistrar) {
	feedv1.RegisterFeedSvcServer(server, f)
}

// CreateFeedEvent 同步调用，数据同步接口
func (f *FeedEventGrpcServer) CreateFeedEvent(ctx context.Context, request *feedv1.CreateFeedEventRequest) (*feedv1.CreateFeedEventResponse, error) {
	err := f.svc.CreateFeedEvent(ctx, f.convertToDomain(request.GetFeedEvent()))
	return &feedv1.CreateFeedEventResponse{}, err
}

func (f *FeedEventGrpcServer) FindFeedEvents(ctx context.Context, request *feedv1.FindFeedEventsRequest) (*feedv1.FindFeedEventsResponse, error) {
	eventList, err := f.svc.GetFeedEventList(ctx, request.GetUid(), request.GetTimestamp(), request.GetLimit())
	if err != nil {
		return &feedv1.FindFeedEventsResponse{}, err
	}
	res := make([]*feedv1.FeedEvent, 0, len(eventList))
	for _, event := range eventList {
		res = append(res, f.convertToView(event))
	}
	return &feedv1.FindFeedEventsResponse{
		FeedEvents: res,
	}, nil
}

func (f *FeedEventGrpcServer) convertToDomain(event *feedv1.FeedEvent) domain.FeedEvent {
	ext := map[string]string{}
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		Id:    event.Id,
		Ctime: time.Unix(event.Ctime, 0),
		Type:  event.Type,
		Ext:   ext,
	}
}

func (f *FeedEventGrpcServer) convertToView(event domain.FeedEvent) *feedv1.FeedEvent {
	val, _ := json.Marshal(event.Ext)
	return &feedv1.FeedEvent{
		Id:      event.Id,
		Type:    event.Type,
		Ctime:   event.Ctime.Unix(),
		Content: string(val),
	}
}
