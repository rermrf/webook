package grpc

import (
	"context"
	"google.golang.org/grpc"
	tagv1 "webook/api/proto/gen/tag/v1"
	"webook/tag/domain"
	"webook/tag/service"
)

type TagServiceServer struct {
	svc service.TagService
	tagv1.UnimplementedTagServiceServer
}

func NewTagServiceServer(svc service.TagService) *TagServiceServer {
	return &TagServiceServer{svc: svc}
}

func (t *TagServiceServer) Register(server *grpc.Server) {
	tagv1.RegisterTagServiceServer(server, t)
}

func (t *TagServiceServer) CreateTag(ctx context.Context, request *tagv1.CreateTagRequest) (*tagv1.CreateTagResponse, error) {
	id, err := t.svc.CreateTag(ctx, request.Uid, request.Name)
	return &tagv1.CreateTagResponse{
		Tag: &tagv1.Tag{
			Id:   id,
			Uid:  request.Uid,
			Name: request.Name,
		},
	}, err
}

func (t *TagServiceServer) AttachTags(ctx context.Context, request *tagv1.AttachTagsRequest) (*tagv1.AttachTagsResponse, error) {
	err := t.svc.AttachTags(ctx, request.Uid, request.Biz, request.BizId, request.Tids)
	return &tagv1.AttachTagsResponse{}, err
}

func (t *TagServiceServer) GetTags(ctx context.Context, request *tagv1.GetTagsRequest) (*tagv1.GetTagsResponse, error) {
	tags, err := t.svc.GetTags(ctx, request.Uid)
	if err != nil {
		return nil, err
	}
	res := make([]*tagv1.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, t.toDTO(tag))
	}
	return &tagv1.GetTagsResponse{
		Tag: res,
	}, nil
}

func (t *TagServiceServer) GetBizTags(ctx context.Context, request *tagv1.GetBizTagsRequest) (*tagv1.GetBizTagsResponse, error) {
	tags, err := t.svc.GetBizTags(ctx, request.Uid, request.Biz, request.BizId)
	if err != nil {
		return nil, err
	}
	res := make([]*tagv1.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, t.toDTO(tag))
	}
	return &tagv1.GetBizTagsResponse{Tags: res}, nil
}

func (t *TagServiceServer) toDTO(tag domain.Tag) *tagv1.Tag {
	return &tagv1.Tag{
		Id:   tag.Id,
		Uid:  tag.Uid,
		Name: tag.Name,
	}
}
