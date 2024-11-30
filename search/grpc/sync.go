package grpc

import (
	"context"
	"google.golang.org/grpc"
	searchv1 "webook/api/proto/gen/search/v1"
	"webook/search/domain"
	"webook/search/service"
)

type SyncServiceServer struct {
	syncSvc service.SyncService
	searchv1.UnimplementedSearchServiceServer
}

func NewSyncServiceServer(syncSvc service.SyncService) *SyncServiceServer {
	return &SyncServiceServer{syncSvc: syncSvc}
}

func (s *SyncServiceServer) Register(server *grpc.Server) {
	searchv1.RegisterSearchServiceServer(server, s)
}

func (s *SyncServiceServer) InputUser(ctx context.Context, request *searchv1.InputUserRequest) (*searchv1.InputUserResponse, error) {
	err := s.syncSvc.InputUser(ctx, s.toDomainUser(request.GetUser()))
	return &searchv1.InputUserResponse{}, err
}

func (s *SyncServiceServer) InputArticle(ctx context.Context, request *searchv1.InputArticleRequest) (*searchv1.InputArticleResponse, error) {
	err := s.syncSvc.InputArticle(ctx, s.toDomainArticle(request.GetArticle()))
	return &searchv1.InputArticleResponse{}, err
}

func (s *SyncServiceServer) InputAny(ctx context.Context, request *searchv1.InputAnyRequest) (*searchv1.InputAnyResponse, error) {
	err := s.syncSvc.InputAny(ctx, request.IndexName, request.DocId, request.Data)
	return &searchv1.InputAnyResponse{}, err
}

func (s *SyncServiceServer) toDomainUser(user *searchv1.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Nickname: user.Nickname,
	}
}

func (s *SyncServiceServer) toDomainArticle(article *searchv1.Article) domain.Article {
	return domain.Article{
		Id:      article.Id,
		Title:   article.Title,
		Status:  article.Status,
		Content: article.Content,
		Tags:    article.Tags,
	}
}
