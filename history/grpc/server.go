package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	historyv1 "webook/api/proto/gen/history/v1"
	"webook/history/domain"
	"webook/history/service"
)

type HistoryServiceServer struct {
	historyv1.UnimplementedHistoryServiceServer
	svc service.HistoryService
}

func NewHistoryServiceServer(svc service.HistoryService) *HistoryServiceServer {
	return &HistoryServiceServer{svc: svc}
}

func (s *HistoryServiceServer) Register(server *grpc.Server) {
	historyv1.RegisterHistoryServiceServer(server, s)
}

func (s *HistoryServiceServer) Record(ctx context.Context, req *historyv1.RecordRequest) (*historyv1.RecordResponse, error) {
	err := s.svc.Record(ctx, domain.BrowseHistory{
		UserId:     req.GetUserId(),
		Biz:        req.GetBiz(),
		BizId:      req.GetBizId(),
		BizTitle:   req.GetBizTitle(),
		AuthorName: req.GetAuthorName(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "记录浏览历史失败: %v", err)
	}
	return &historyv1.RecordResponse{}, nil
}

func (s *HistoryServiceServer) List(ctx context.Context, req *historyv1.ListRequest) (*historyv1.ListResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	records, hasMore, err := s.svc.List(ctx, req.GetUserId(), req.GetCursor(), limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询浏览历史失败: %v", err)
	}
	items := make([]*historyv1.HistoryItem, 0, len(records))
	for _, r := range records {
		items = append(items, &historyv1.HistoryItem{
			Id:         r.Id,
			Biz:        r.Biz,
			BizId:      r.BizId,
			BizTitle:   r.BizTitle,
			AuthorName: r.AuthorName,
			Ctime:      r.Ctime,
			Utime:      r.Utime,
		})
	}
	return &historyv1.ListResponse{
		Items:   items,
		HasMore: hasMore,
	}, nil
}

func (s *HistoryServiceServer) Clear(ctx context.Context, req *historyv1.ClearRequest) (*historyv1.ClearResponse, error) {
	err := s.svc.Clear(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "清除浏览历史失败: %v", err)
	}
	return &historyv1.ClearResponse{}, nil
}
