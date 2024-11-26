package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
	commentv1 "webook/api/proto/gen/comment/v1"
	"webook/comment/domain"
	"webook/comment/service"
)

type CommentServiceServer struct {
	commentv1.UnimplementedCommentServiceServer
	svc service.CommentService
}

func NewCommentServiceServer(svc service.CommentService) *CommentServiceServer {
	return &CommentServiceServer{svc: svc}
}

func (c *CommentServiceServer) Register(server *grpc.Server) {
	commentv1.RegisterCommentServiceServer(server, c)
}

func (c *CommentServiceServer) GetCommentList(ctx context.Context, request *commentv1.GetCommentListRequest) (*commentv1.GetCommentListResponse, error) {
	minID := request.GetMinId()
	// 第一次查询，用户没有传
	if minID <= 0 {
		// 从当前最新的评论开始取
		minID = math.MaxInt64
	}
	domainComments, err := c.svc.GetCommentList(ctx, request.GetBiz(), request.GetBizid(), request.GetMinId(), request.GetLimit())
	if err != nil {
		return nil, err
	}
	return &commentv1.GetCommentListResponse{
		Comments: c.toDTO(domainComments),
	}, nil
}

func (c *CommentServiceServer) DeleteComment(ctx context.Context, request *commentv1.DeleteCommentRequest) (*commentv1.DeleteCommentResponse, error) {
	err := c.svc.DeleteComment(ctx, request.Id)
	return &commentv1.DeleteCommentResponse{}, err
}

// CreateComment 传入父 parent 的id，那么就是代表回复了某个评论
func (c *CommentServiceServer) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	err := c.svc.CreateComment(ctx, c.convertToDomain(request.GetComment()))
	return &commentv1.CreateCommentResponse{}, err
}

func (c *CommentServiceServer) GetMoreReplies(ctx context.Context, request *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	cs, err := c.svc.GetMoreReplies(ctx, request.Rid, request.MinId, request.Limit)
	if err != nil {
		return nil, err
	}
	return &commentv1.GetMoreRepliesResponse{
		Replies: c.toDTO(cs),
	}, nil
}

func (c *CommentServiceServer) toDTO(dcs []domain.Comment) []*commentv1.Comment {
	rpcComments := make([]*commentv1.Comment, 0, len(dcs))
	for _, dc := range dcs {
		rc := &commentv1.Comment{
			Id:      dc.Id,
			Biz:     dc.Biz,
			Bizid:   dc.BizId,
			Uid:     dc.Commentator.ID,
			Content: dc.Content,
			Ctime:   timestamppb.New(dc.Ctime),
			Utime:   timestamppb.New(dc.Utime),
		}
		if dc.RootComment != nil {
			rc.RootComment = &commentv1.Comment{
				Id: dc.RootComment.Id,
			}
		}
		if dc.ParentComment != nil {
			rc.ParentComment = &commentv1.Comment{
				Id: dc.ParentComment.Id,
			}
		}
		rpcComments = append(rpcComments, rc)
	}
	rpcCommentMap := make(map[int64]*commentv1.Comment, len(rpcComments))
	for _, rpcomment := range rpcComments {
		rpcCommentMap[rpcomment.Id] = rpcomment
	}
	for _, dc := range dcs {
		rpcComment := rpcCommentMap[dc.Id]
		if dc.RootComment != nil {
			val, ok := rpcCommentMap[dc.RootComment.Id]
			if !ok {
				rpcComment.RootComment = val
			}
		}
		if dc.ParentComment != nil {
			val, ok := rpcCommentMap[dc.ParentComment.Id]
			if !ok {
				rpcComment.ParentComment = val
			}
		}
	}
	return rpcComments
}

func (c *CommentServiceServer) convertToDomain(comment *commentv1.Comment) domain.Comment {
	dc := domain.Comment{
		Id:      comment.Id,
		Biz:     comment.Biz,
		BizId:   comment.Bizid,
		Content: comment.Content,
		Commentator: domain.User{
			ID: comment.Uid,
		},
	}
	if comment.GetParentComment() != nil {
		dc.ParentComment = &domain.Comment{
			Id: comment.GetParentComment().Id,
		}
	}
	if comment.GetRootComment() != nil {
		dc.RootComment = &domain.Comment{
			Id: comment.GetRootComment().Id,
		}
	}
	return dc
}
