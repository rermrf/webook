package grpc

import (
	"context"
	commentv1 "webook/api/proto/gen/comment/v1"
)

type CommentServiceAsyncServer struct {
	CommentServiceServer
}

func (c *CommentServiceAsyncServer) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	if ctx.Value("limited") == "true" || ctx.Value("downgrad") == "true" {
		// 在这里发送到 Kafka 里面
		return &commentv1.CreateCommentResponse{}, nil
	} else {
		err := c.svc.CreateComment(ctx, c.convertToDomain(request.GetComment()))
		return &commentv1.CreateCommentResponse{}, err
	}
}
