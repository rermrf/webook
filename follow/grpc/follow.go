package grpc

import (
	"context"
	"google.golang.org/grpc"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/follow/domain"
	"webook/follow/service"
)

type FollowServiceServer struct {
	svc service.FollowService
	followv1.UnimplementedFollowServiceServer
}

func NewFollowServiceServer(svc service.FollowService) *FollowServiceServer {
	return &FollowServiceServer{svc: svc}
}

func (f *FollowServiceServer) Register(server *grpc.Server) {
	followv1.RegisterFollowServiceServer(server, f)
}

// Follow 关注
func (f *FollowServiceServer) Follow(ctx context.Context, request *followv1.FollowRequest) (*followv1.FollowResponse, error) {
	err := f.svc.Follow(ctx, request.Follower, request.Followee)
	return &followv1.FollowResponse{}, err
}

// CancelFollow 取消关注
func (f *FollowServiceServer) CancelFollow(ctx context.Context, request *followv1.CancelFollowRequest) (*followv1.CancelFollowResponse, error) {
	err := f.svc.CancelFollow(ctx, request.Follower, request.Followee)
	return &followv1.CancelFollowResponse{}, err
}

// GetFollowee 获取关注者
func (f *FollowServiceServer) GetFollowee(ctx context.Context, request *followv1.GetFolloweeRequest) (*followv1.GetFolloweeResponse, error) {
	list, err := f.svc.GetFollowee(ctx, request.Follower, request.Offset, request.Limit)
	if err != nil {
		return nil, err
	}
	res := make([]*followv1.FollowRelation, 0, len(list))
	for _, v := range list {
		res = append(res, f.convertToView(v))
	}
	return &followv1.GetFolloweeResponse{FollowRelations: res}, nil
}

// FollowInfo 获取关注关系
func (f *FollowServiceServer) FollowInfo(ctx context.Context, request *followv1.FollowInfoRequest) (*followv1.FollowInfoResponse, error) {
	res, err := f.svc.FollowInfo(ctx, request.Follower, request.Followee)
	if err != nil {
		return nil, err
	}
	return &followv1.FollowInfoResponse{FollowRelation: f.convertToView(res)}, nil
}

// GetFollower 获取粉丝
func (f *FollowServiceServer) GetFollower(ctx context.Context, request *followv1.GetFollowerRequest) (*followv1.GetFollowerResponse, error) {
	list, err := f.svc.GetFollower(ctx, request.Followee, request.Offset, request.Limit)
	res := make([]*followv1.FollowRelation, 0, len(list))
	for _, v := range list {
		res = append(res, f.convertToView(v))
	}
	return &followv1.GetFollowerResponse{FollowRelations: res}, err
}

func (f *FollowServiceServer) GetFollowStatic(ctx context.Context, request *followv1.GetFollowStaticRequest) (*followv1.GetFollowStaticResponse, error) {
	res, err := f.svc.GetFollowStatic(ctx, request.Followee)
	return &followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followers: res.Followers,
			Followees: res.Followees,
		},
	}, err
}

func (f *FollowServiceServer) convertToView(src domain.FollowRelation) *followv1.FollowRelation {
	return &followv1.FollowRelation{
		Follower: src.Follower,
		Followee: src.Followee,
	}
}
