package grpc

import (
	"context"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/follow/service"
)

type FollowServiceServer struct {
	svc service.FollowService
	followv1.UnimplementedFollowServiceServer
}

func (f FollowServiceServer) Follow(ctx context.Context, request *followv1.FollowRequest) (*followv1.FollowResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f FollowServiceServer) CancelFollow(ctx context.Context, request *followv1.CancelFollowRequest) (*followv1.CancelFollowResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f FollowServiceServer) GetFollowee(ctx context.Context, request *followv1.GetFolloweeRequest) (*followv1.GetFolloweeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f FollowServiceServer) FollowInfo(ctx context.Context, request *followv1.FollowInfoRequest) (*followv1.FollowInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}
