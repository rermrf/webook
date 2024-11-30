package grpc

import (
	"context"
	"google.golang.org/grpc"
	searchv1 "webook/api/proto/gen/search/v1"
	"webook/search/service"
)

type SearchServiceServer struct {
	svc service.SearchService
	searchv1.UnimplementedSearchServiceServer
}

func NewSearchServiceServer(svc service.SearchService) *SearchServiceServer {
	return &SearchServiceServer{svc: svc}
}

func (s *SearchServiceServer) Register(server *grpc.Server) {
	searchv1.RegisterSearchServiceServer(server, s)
}

func (s *SearchServiceServer) Search(ctx context.Context, request *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
	resp, err := s.svc.Search(ctx, request.Uid, request.Expression)
	if err != nil {
		return nil, err
	}
	users := make([]*searchv1.User, 0, len(resp.Users))
	for _, user := range resp.Users {
		users = append(users, &searchv1.User{
			Id:       user.Id,
			Nickname: user.Nickname,
			Email:    user.Email,
			Phone:    user.Phone,
		})
	}
	articles := make([]*searchv1.Article, 0, len(resp.Articles))
	for _, article := range resp.Articles {
		articles = append(articles, &searchv1.Article{
			Id:      article.Id,
			Title:   article.Title,
			Status:  article.Status,
			Content: article.Content,
		})
	}
	return &searchv1.SearchResponse{
		User: &searchv1.UserResult{
			Users: users,
		},
		Article: &searchv1.ArticleResult{
			Articles: articles,
		},
	}, nil
}
