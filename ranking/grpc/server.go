package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	rankingv1 "webook/api/proto/gen/ranking/v1"
	"webook/ranking/service"
)

type RankingServiceServer struct {
	svc service.RankingService
	rankingv1.UnimplementedRankingServiceServer
}

func NewRankingServiceServer(svc service.RankingService) *RankingServiceServer {
	return &RankingServiceServer{svc: svc}
}

func (r *RankingServiceServer) Register(server *grpc.Server) {
	rankingv1.RegisterRankingServiceServer(server, r)
}

func (r *RankingServiceServer) TopN(ctx context.Context, request *rankingv1.TopNRequest) (*rankingv1.TopNResponse, error) {
	arts, err := r.svc.TopN(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]*rankingv1.Article, len(arts))
	for i, art := range arts {
		resp[i] = &rankingv1.Article{
			Id:      art.Id,
			Title:   art.Title,
			Content: art.Content,
			Author: &rankingv1.Author{
				Id:   art.Author.Id,
				Name: art.Author.Name,
			},
			Status: int32(art.Status),
			Ctime:  timestamppb.New(art.Ctime),
			Utime:  timestamppb.New(art.Utime),
		}
	}
	return &rankingv1.TopNResponse{
		Articles: resp,
	}, nil
}

func (r *RankingServiceServer) RankTopN(ctx context.Context, request *rankingv1.RankTopNRequest) (*rankingv1.RankTopNResponse, error) {
	err := r.svc.RankTopN(ctx)
	return &rankingv1.RankTopNResponse{}, err
}
