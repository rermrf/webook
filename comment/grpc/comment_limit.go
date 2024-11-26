package grpc

import (
	"context"
	"errors"
	commentv1 "webook/api/proto/gen/comment/v1"
)

// RateLimitCommentServiceServer 容错，当系统降级时，非热门资源不进入
type RateLimitCommentServiceServer struct {
	CommentServiceServer
}

func (c *RateLimitCommentServiceServer) GetCommentList(ctx context.Context, request *commentv1.GetCommentListRequest) (*commentv1.GetCommentListResponse, error) {
	if ctx.Value("downgrade") == "true" && !c.hotBiz(request.Biz, request.GetBizid()) {
		return nil, errors.New("触发了降级，非热门资源")
	}
	return c.CommentServiceServer.GetCommentList(ctx, request)
}

func (c *RateLimitCommentServiceServer) hotBiz(biz string, bizId int64) bool {
	// 这个热门资源怎么判定
	// 一般是借助周期性的任务来计算一个白名单，放到 redis 中
	return true
}

func (c *RateLimitCommentServiceServer) GetMoreReplies(ctx context.Context, request *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	if ctx.Value("downgrade") == "true" {
		return nil, errors.New("触发了降级")
	}
	return c.CommentServiceServer.GetMoreReplies(ctx, request)
}
