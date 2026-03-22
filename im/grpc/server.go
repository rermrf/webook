package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	imv1 "webook/api/proto/gen/im/v1"
	"webook/im/domain"
	"webook/im/service"
)

type IMServiceServer struct {
	imv1.UnimplementedIMServiceServer
	msgSvc  service.MessageService
	convSvc service.ConversationService
}

func NewIMServiceServer(msgSvc service.MessageService, convSvc service.ConversationService) *IMServiceServer {
	return &IMServiceServer{
		msgSvc:  msgSvc,
		convSvc: convSvc,
	}
}

func (s *IMServiceServer) Register(server *grpc.Server) {
	imv1.RegisterIMServiceServer(server, s)
}

func (s *IMServiceServer) SendMessage(ctx context.Context, req *imv1.SendMessageRequest) (*imv1.SendMessageResponse, error) {
	msg := domain.Message{
		SenderId:   req.GetSenderId(),
		ReceiverId: req.GetReceiverId(),
		MsgType:    domain.MsgType(req.GetMsgType()),
		Content:    req.GetContent(),
	}
	result, err := s.msgSvc.SendMessage(ctx, msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "发送消息失败: %v", err)
	}
	return &imv1.SendMessageResponse{
		MessageId:      result.Id,
		ConversationId: result.ConversationID,
		Ctime:          result.Ctime,
	}, nil
}

func (s *IMServiceServer) ListMessages(ctx context.Context, req *imv1.ListMessagesRequest) (*imv1.ListMessagesResponse, error) {
	messages, hasMore, err := s.msgSvc.ListMessages(ctx, req.GetConversationId(), req.GetCursor(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询消息失败: %v", err)
	}
	items := make([]*imv1.MessageItem, 0, len(messages))
	for _, m := range messages {
		items = append(items, toMessageItem(m))
	}
	return &imv1.ListMessagesResponse{
		Messages: items,
		HasMore:  hasMore,
	}, nil
}

func (s *IMServiceServer) MarkAsRead(ctx context.Context, req *imv1.MarkAsReadRequest) (*imv1.MarkAsReadResponse, error) {
	err := s.msgSvc.MarkAsRead(ctx, req.GetUserId(), req.GetConversationId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "标记已读失败: %v", err)
	}
	return &imv1.MarkAsReadResponse{}, nil
}

func (s *IMServiceServer) RecallMessage(ctx context.Context, req *imv1.RecallMessageRequest) (*imv1.RecallMessageResponse, error) {
	err := s.msgSvc.RecallMessage(ctx, req.GetUserId(), req.GetMessageId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "撤回消息失败: %v", err)
	}
	return &imv1.RecallMessageResponse{}, nil
}

func (s *IMServiceServer) ListConversations(ctx context.Context, req *imv1.ListConversationsRequest) (*imv1.ListConversationsResponse, error) {
	conversations, hasMore, err := s.convSvc.ListConversations(ctx, req.GetUserId(), req.GetCursor(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询会话列表失败: %v", err)
	}
	items := make([]*imv1.ConversationItem, 0, len(conversations))
	for _, c := range conversations {
		items = append(items, toConversationItem(c))
	}
	return &imv1.ListConversationsResponse{
		Conversations: items,
		HasMore:       hasMore,
	}, nil
}

func (s *IMServiceServer) GetConversation(ctx context.Context, req *imv1.GetConversationRequest) (*imv1.GetConversationResponse, error) {
	conv, err := s.convSvc.GetConversation(ctx, req.GetConversationId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询会话失败: %v", err)
	}
	return &imv1.GetConversationResponse{
		Conversation: toConversationItem(conv),
	}, nil
}

func (s *IMServiceServer) GetUnreadCount(ctx context.Context, req *imv1.GetUnreadCountRequest) (*imv1.GetUnreadCountResponse, error) {
	total, byConversation, err := s.convSvc.GetUnreadCount(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询未读计数失败: %v", err)
	}
	return &imv1.GetUnreadCountResponse{
		Total:          total,
		ByConversation: byConversation,
	}, nil
}

func (s *IMServiceServer) IsOnline(ctx context.Context, req *imv1.IsOnlineRequest) (*imv1.IsOnlineResponse, error) {
	online, err := s.convSvc.IsOnline(ctx, req.GetUserId())
	if err != nil {
		return &imv1.IsOnlineResponse{Online: false}, nil
	}
	return &imv1.IsOnlineResponse{Online: online}, nil
}

func toMessageItem(m domain.Message) *imv1.MessageItem {
	return &imv1.MessageItem{
		Id:             m.Id,
		ConversationId: m.ConversationID,
		SenderId:       m.SenderId,
		ReceiverId:     m.ReceiverId,
		MsgType:        uint32(m.MsgType),
		Content:        m.Content,
		Status:         uint32(m.Status),
		Ctime:          m.Ctime,
	}
}

func toConversationItem(c domain.Conversation) *imv1.ConversationItem {
	item := &imv1.ConversationItem{
		ConversationId: c.ConversationID,
		Members:        c.Members,
		Utime:          c.Utime,
	}
	if c.LastMsg.Ctime > 0 {
		item.LastMsg = &imv1.MessageItem{
			SenderId: c.LastMsg.SenderId,
			MsgType:  uint32(c.LastMsg.MsgType),
			Content:  c.LastMsg.Content,
			Ctime:    c.LastMsg.Ctime,
		}
	}
	return item
}
