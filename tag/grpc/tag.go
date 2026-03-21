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
	id, err := t.svc.CreateTag(ctx, request.Name, request.Description)
	return &tagv1.CreateTagResponse{
		Tag: &tagv1.Tag{
			Id:          id,
			Name:        request.Name,
			Description: request.Description,
		},
	}, err
}

func (t *TagServiceServer) AttachTags(ctx context.Context, request *tagv1.AttachTagsRequest) (*tagv1.AttachTagsResponse, error) {
	err := t.svc.AttachTags(ctx, request.Biz, request.BizId, request.Tids)
	return &tagv1.AttachTagsResponse{}, err
}

func (t *TagServiceServer) GetTags(ctx context.Context, request *tagv1.GetTagsRequest) (*tagv1.GetTagsResponse, error) {
	tags, err := t.svc.GetTags(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*tagv1.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, t.toDTO(tag))
	}
	return &tagv1.GetTagsResponse{Tags: res}, nil
}

func (t *TagServiceServer) GetTagById(ctx context.Context, request *tagv1.GetTagByIdRequest) (*tagv1.GetTagByIdResponse, error) {
	tag, err := t.svc.GetTagById(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &tagv1.GetTagByIdResponse{Tag: t.toDTO(tag)}, nil
}

func (t *TagServiceServer) GetBizTags(ctx context.Context, request *tagv1.GetBizTagsRequest) (*tagv1.GetBizTagsResponse, error) {
	tags, err := t.svc.GetBizTags(ctx, request.Biz, request.BizId)
	if err != nil {
		return nil, err
	}
	res := make([]*tagv1.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, t.toDTO(tag))
	}
	return &tagv1.GetBizTagsResponse{Tags: res}, nil
}

func (t *TagServiceServer) GetBizIdsByTag(ctx context.Context, request *tagv1.GetBizIdsByTagRequest) (*tagv1.GetBizIdsByTagResponse, error) {
	ids, err := t.svc.GetBizIdsByTag(ctx, request.Biz, request.TagId, int(request.Offset), int(request.Limit), request.SortBy)
	if err != nil {
		return nil, err
	}
	return &tagv1.GetBizIdsByTagResponse{BizIds: ids}, nil
}

func (t *TagServiceServer) CountBizByTag(ctx context.Context, request *tagv1.CountBizByTagRequest) (*tagv1.CountBizByTagResponse, error) {
	count, err := t.svc.CountBizByTag(ctx, request.Biz, request.TagId)
	if err != nil {
		return nil, err
	}
	return &tagv1.CountBizByTagResponse{Count: count}, nil
}

func (t *TagServiceServer) toDTO(tag domain.Tag) *tagv1.Tag {
	return &tagv1.Tag{
		Id:            tag.Id,
		Name:          tag.Name,
		Description:   tag.Description,
		FollowerCount: tag.FollowerCount,
	}
}

func (t *TagServiceServer) FollowTag(ctx context.Context, request *tagv1.FollowTagRequest) (*tagv1.FollowTagResponse, error) {
	err := t.svc.FollowTag(ctx, request.Uid, request.TagId)
	return &tagv1.FollowTagResponse{}, err
}

func (t *TagServiceServer) UnfollowTag(ctx context.Context, request *tagv1.UnfollowTagRequest) (*tagv1.UnfollowTagResponse, error) {
	err := t.svc.UnfollowTag(ctx, request.Uid, request.TagId)
	return &tagv1.UnfollowTagResponse{}, err
}

func (t *TagServiceServer) CheckTagFollow(ctx context.Context, request *tagv1.CheckTagFollowRequest) (*tagv1.CheckTagFollowResponse, error) {
	followed, err := t.svc.CheckTagFollow(ctx, request.Uid, request.TagId)
	if err != nil {
		return nil, err
	}
	return &tagv1.CheckTagFollowResponse{Followed: followed}, nil
}

func (t *TagServiceServer) GetUserFollowedTags(ctx context.Context, request *tagv1.GetUserFollowedTagsRequest) (*tagv1.GetUserFollowedTagsResponse, error) {
	tags, err := t.svc.GetUserFollowedTags(ctx, request.Uid, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, err
	}
	res := make([]*tagv1.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, t.toDTO(tag))
	}
	return &tagv1.GetUserFollowedTagsResponse{Tags: res}, nil
}

func (t *TagServiceServer) BatchGetBizTags(ctx context.Context, request *tagv1.BatchGetBizTagsRequest) (*tagv1.BatchGetBizTagsResponse, error) {
	tagMap, err := t.svc.BatchGetBizTags(ctx, request.Biz, request.BizIds)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]*tagv1.BizTagList, len(tagMap))
	for bizId, tags := range tagMap {
		list := make([]*tagv1.Tag, 0, len(tags))
		for _, tag := range tags {
			list = append(list, t.toDTO(tag))
		}
		result[bizId] = &tagv1.BizTagList{Tags: list}
	}
	return &tagv1.BatchGetBizTagsResponse{BizTags: result}, nil
}
