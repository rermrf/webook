package grpc

import (
	"context"
	"google.golang.org/grpc"
	articlev1 "webook/api/proto/gen/article/v1"
	"webook/article/domain"
	"webook/article/service"
)

type ArticleGRPCServer struct {
	svc service.ArticleService
	articlev1.UnimplementedArticleServiceServer
}

func NewArticleGRPCServer(svc service.ArticleService) *ArticleGRPCServer {
	return &ArticleGRPCServer{svc: svc}
}

func (a *ArticleGRPCServer) Register(server *grpc.Server) {
	articlev1.RegisterArticleServiceServer(server, a)
}

func (a *ArticleGRPCServer) Save(ctx context.Context, request *articlev1.SaveRequest) (*articlev1.SaveResponse, error) {
	id, err := a.svc.Save(ctx, a.toDTO(request.GetArticle()))
	return &articlev1.SaveResponse{Id: id}, err
}

func (a *ArticleGRPCServer) Publish(ctx context.Context, request *articlev1.PublishRequest) (*articlev1.PublishResponse, error) {
	id, err := a.svc.Publish(ctx, a.toDTO(request.GetArticle()))
	return &articlev1.PublishResponse{
		Id: id,
	}, err
}

func (a *ArticleGRPCServer) WithDraw(ctx context.Context, request *articlev1.WithDrawRequest) (*articlev1.WithDrawResponse, error) {
	err := a.svc.WithDraw(ctx, a.toDTO(request.GetArticle()))
	return &articlev1.WithDrawResponse{}, err
}

func (a *ArticleGRPCServer) List(ctx context.Context, request *articlev1.ListRequest) (*articlev1.ListResponse, error) {
	arts, err := a.svc.List(ctx, request.GetUid(), int(request.GetOffset()), int(request.GetLimit()))
	res := make([]*articlev1.Article, len(arts))
	for i, v := range arts {
		res[i] = a.toV(v)
	}
	return &articlev1.ListResponse{Articles: res}, err
}

func (a *ArticleGRPCServer) GetById(ctx context.Context, request *articlev1.GetByIdRequest) (*articlev1.GetByIdResponse, error) {
	art, err := a.svc.GetById(ctx, request.GetId())
	return &articlev1.GetByIdResponse{Article: a.toV(art)}, err
}

func (a *ArticleGRPCServer) GetPublishedById(ctx context.Context, request *articlev1.GetPublishedByIdRequest) (*articlev1.GetPublishedByIdResponse, error) {
	art, err := a.svc.GetPublishedById(ctx, request.GetId(), request.GetUid())
	return &articlev1.GetPublishedByIdResponse{Article: a.toV(art)}, err
}

func (a *ArticleGRPCServer) ListPub(ctx context.Context, request *articlev1.ListPubRequest) (*articlev1.ListPubResponse, error) {
	arts, err := a.svc.ListPub(ctx, request.GetStartTime().AsTime(), int(request.GetOffset()), int(request.GetLimit()))
	res := make([]*articlev1.Article, len(arts))
	for i, v := range arts {
		res[i] = a.toV(v)
	}
	return &articlev1.ListPubResponse{Articles: res}, err
}

func (a *ArticleGRPCServer) toDTO(art *articlev1.Article) domain.Article {
	return domain.Article{
		Id:      art.GetId(),
		Title:   art.GetTitle(),
		Content: art.GetContent(),
		Author: domain.Author{
			Id:   art.GetAuthor().GetId(),
			Name: art.GetAuthor().GetName(),
		},
		Status: domain.ArticleStatus(art.GetStatus()),
	}
}

func (a *ArticleGRPCServer) toV(art domain.Article) *articlev1.Article {
	return &articlev1.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: &articlev1.Author{
			Id:   art.Author.Id,
			Name: art.Author.Name,
		},
		Status: int32(art.Status.ToUint8()),
	}
}
