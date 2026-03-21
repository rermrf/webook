package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notificationv2 "webook/api/proto/gen/notification/v2"
	"webook/notification/domain"
	"webook/notification/service"
)

type NotificationServiceServer struct {
	notificationv2.UnimplementedNotificationServiceServer
	svc    service.NotificationService
	tplSvc service.TemplateService
}

func NewNotificationServiceServer(svc service.NotificationService, tplSvc service.TemplateService) *NotificationServiceServer {
	return &NotificationServiceServer{svc: svc, tplSvc: tplSvc}
}

func (s *NotificationServiceServer) Register(server *grpc.Server) {
	notificationv2.RegisterNotificationServiceServer(server, s)
}

// Send 发送通知
func (s *NotificationServiceServer) Send(ctx context.Context, req *notificationv2.SendRequest) (*notificationv2.SendResponse, error) {
	n := req.GetNotification()
	if n == nil {
		return nil, status.Error(codes.InvalidArgument, "notification is required")
	}
	dn := s.toDomain(n)
	_, err := s.svc.Send(ctx, dn)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "发送通知失败: %v", err)
	}
	return &notificationv2.SendResponse{}, nil
}

// BatchSend 批量发送通知
func (s *NotificationServiceServer) BatchSend(ctx context.Context, req *notificationv2.BatchSendRequest) (*notificationv2.BatchSendResponse, error) {
	protoNotifications := req.GetNotifications()
	if len(protoNotifications) == 0 {
		return &notificationv2.BatchSendResponse{}, nil
	}
	// 将每个 proto Notification 的 receivers 拆分为单独的 domain.Notification
	var domainNotifications []domain.Notification
	for _, pn := range protoNotifications {
		receivers := pn.GetReceivers()
		if len(receivers) == 0 {
			// 如果 receivers 为空，使用 receiver 字段
			domainNotifications = append(domainNotifications, s.toDomain(pn))
		} else {
			for _, receiver := range receivers {
				dn := s.toDomain(pn)
				dn.Receiver = receiver
				domainNotifications = append(domainNotifications, dn)
			}
		}
	}
	_, err := s.svc.BatchSend(ctx, domainNotifications)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "批量发送通知失败: %v", err)
	}
	return &notificationv2.BatchSendResponse{}, nil
}

// Prepare 事务预提交
func (s *NotificationServiceServer) Prepare(ctx context.Context, req *notificationv2.PrepareRequest) (*notificationv2.PrepareResponse, error) {
	n := req.GetNotification()
	if n == nil {
		return nil, status.Error(codes.InvalidArgument, "notification is required")
	}
	domainReq := domain.PrepareRequest{
		Notification:       s.toDomain(n),
		BizId:              req.GetBizId(),
		CheckBackTimeoutMs: req.GetCheckBackTimeoutMs(),
	}
	_, _, err := s.svc.Prepare(ctx, domainReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "事务预提交失败: %v", err)
	}
	return &notificationv2.PrepareResponse{}, nil
}

// Confirm 事务确认
func (s *NotificationServiceServer) Confirm(ctx context.Context, req *notificationv2.ConfirmRequest) (*notificationv2.ConfirmResponse, error) {
	err := s.svc.Confirm(ctx, req.GetKey())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "事务确认失败: %v", err)
	}
	return &notificationv2.ConfirmResponse{}, nil
}

// Cancel 事务取消
func (s *NotificationServiceServer) Cancel(ctx context.Context, req *notificationv2.CancelRequest) (*notificationv2.CancelResponse, error) {
	err := s.svc.Cancel(ctx, req.GetKey())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "事务取消失败: %v", err)
	}
	return &notificationv2.CancelResponse{}, nil
}

// CreateTemplate 创建模板
func (s *NotificationServiceServer) CreateTemplate(ctx context.Context, req *notificationv2.CreateTemplateRequest) (*notificationv2.CreateTemplateResponse, error) {
	tpl := req.GetTemplate()
	if tpl == nil {
		return nil, status.Error(codes.InvalidArgument, "template is required")
	}
	domainTpl := s.toDomainTemplate(tpl)
	id, err := s.tplSvc.Create(ctx, domainTpl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建模板失败: %v", err)
	}
	return &notificationv2.CreateTemplateResponse{Id: id}, nil
}

// UpdateTemplate 更新模板
func (s *NotificationServiceServer) UpdateTemplate(ctx context.Context, req *notificationv2.UpdateTemplateRequest) (*notificationv2.UpdateTemplateResponse, error) {
	tpl := req.GetTemplate()
	if tpl == nil {
		return nil, status.Error(codes.InvalidArgument, "template is required")
	}
	domainTpl := s.toDomainTemplate(tpl)
	err := s.tplSvc.Update(ctx, domainTpl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "更新模板失败: %v", err)
	}
	return &notificationv2.UpdateTemplateResponse{}, nil
}

// GetTemplate 获取模板
func (s *NotificationServiceServer) GetTemplate(ctx context.Context, req *notificationv2.GetTemplateRequest) (*notificationv2.GetTemplateResponse, error) {
	tpl, err := s.tplSvc.GetByTemplateId(ctx, req.GetTemplateId(), domain.ChannelInApp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取模板失败: %v", err)
	}
	return &notificationv2.GetTemplateResponse{
		Template: s.toProtoTemplate(tpl),
	}, nil
}

// ListTemplates 模板列表
func (s *NotificationServiceServer) ListTemplates(ctx context.Context, req *notificationv2.ListTemplatesRequest) (*notificationv2.ListTemplatesResponse, error) {
	templates, err := s.tplSvc.List(ctx, domain.ChannelInApp, int(req.GetOffset()), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询模板列表失败: %v", err)
	}
	protoTemplates := make([]*notificationv2.Template, 0, len(templates))
	for _, tpl := range templates {
		protoTemplates = append(protoTemplates, s.toProtoTemplate(tpl))
	}
	return &notificationv2.ListTemplatesResponse{
		Templates: protoTemplates,
	}, nil
}

// ListNotifications 通知列表
func (s *NotificationServiceServer) ListNotifications(ctx context.Context, req *notificationv2.ListNotificationsRequest) (*notificationv2.ListNotificationsResponse, error) {
	notifications, err := s.svc.List(ctx, req.GetUserId(), int(req.GetOffset()), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询通知列表失败: %v", err)
	}
	return &notificationv2.ListNotificationsResponse{
		Notifications: s.toNotificationItems(notifications),
	}, nil
}

// ListByGroup 按分组获取通知列表
func (s *NotificationServiceServer) ListByGroup(ctx context.Context, req *notificationv2.ListByGroupRequest) (*notificationv2.ListByGroupResponse, error) {
	group := domain.NotificationGroup(req.GetGroup())
	notifications, err := s.svc.ListByGroup(ctx, req.GetUserId(), group, int(req.GetOffset()), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "按分组查询通知失败: %v", err)
	}
	return &notificationv2.ListByGroupResponse{
		Notifications: s.toNotificationItems(notifications),
	}, nil
}

// ListUnread 获取未读通知列表
func (s *NotificationServiceServer) ListUnread(ctx context.Context, req *notificationv2.ListUnreadRequest) (*notificationv2.ListUnreadResponse, error) {
	notifications, err := s.svc.ListUnread(ctx, req.GetUserId(), int(req.GetOffset()), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询未读通知失败: %v", err)
	}
	return &notificationv2.ListUnreadResponse{
		Notifications: s.toNotificationItems(notifications),
	}, nil
}

// MarkAsRead 标记单条通知已读
func (s *NotificationServiceServer) MarkAsRead(ctx context.Context, req *notificationv2.MarkAsReadRequest) (*notificationv2.MarkAsReadResponse, error) {
	err := s.svc.MarkAsRead(ctx, req.GetUserId(), []int64{req.GetId()})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "标记已读失败: %v", err)
	}
	return &notificationv2.MarkAsReadResponse{}, nil
}

// MarkAllAsRead 标记全部通知已读
func (s *NotificationServiceServer) MarkAllAsRead(ctx context.Context, req *notificationv2.MarkAllAsReadRequest) (*notificationv2.MarkAllAsReadResponse, error) {
	err := s.svc.MarkAllAsRead(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "标记全部已读失败: %v", err)
	}
	return &notificationv2.MarkAllAsReadResponse{}, nil
}

// GetUnreadCount 获取未读数统计
func (s *NotificationServiceServer) GetUnreadCount(ctx context.Context, req *notificationv2.GetUnreadCountRequest) (*notificationv2.GetUnreadCountResponse, error) {
	count, err := s.svc.GetUnreadCount(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取未读数统计失败: %v", err)
	}
	byGroup := make(map[int32]int64, len(count.ByGroup))
	for g, c := range count.ByGroup {
		byGroup[int32(g)] = c
	}
	return &notificationv2.GetUnreadCountResponse{
		Total:   count.Total,
		ByGroup: byGroup,
	}, nil
}

// Delete 删除通知
func (s *NotificationServiceServer) Delete(ctx context.Context, req *notificationv2.DeleteRequest) (*notificationv2.DeleteResponse, error) {
	err := s.svc.Delete(ctx, req.GetUserId(), req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "删除通知失败: %v", err)
	}
	return &notificationv2.DeleteResponse{}, nil
}

// DeleteAll 删除所有通知
func (s *NotificationServiceServer) DeleteAll(ctx context.Context, req *notificationv2.DeleteAllRequest) (*notificationv2.DeleteAllResponse, error) {
	err := s.svc.DeleteAll(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "删除所有通知失败: %v", err)
	}
	return &notificationv2.DeleteAllResponse{}, nil
}

// toDomain 将 proto Notification 转换为 domain.Notification
func (s *NotificationServiceServer) toDomain(n *notificationv2.Notification) domain.Notification {
	return domain.Notification{
		Key:            n.GetKey(),
		Channel:        domain.Channel(n.GetChannel()),
		Receiver:       n.GetReceiver(),
		TemplateId:     n.GetTemplateId(),
		TemplateParams: n.GetTemplateParams(),
		Strategy:       domain.SendStrategy(n.GetStrategy()),
		ScheduledTime:  n.GetScheduledTime(),
		GroupType:      domain.NotificationGroup(n.GetGroupType()),
		SourceId:       n.GetSourceId(),
		SourceName:     n.GetSourceName(),
		TargetId:       n.GetTargetId(),
		TargetType:     n.GetTargetType(),
		TargetTitle:    n.GetTargetTitle(),
	}
}

// toNotificationItem 将 domain.Notification 转换为 proto NotificationItem
func (s *NotificationServiceServer) toNotificationItem(n domain.Notification) *notificationv2.NotificationItem {
	return &notificationv2.NotificationItem{
		Id:         n.Id,
		GroupType:  notificationv2.NotificationGroup(n.GroupType),
		SourceId:   n.SourceId,
		SourceName: n.SourceName,
		TargetId:   n.TargetId,
		TargetType: n.TargetType,
		TargetTitle: n.TargetTitle,
		Content:    n.Content,
		IsRead:     n.IsRead,
		Ctime:      n.Ctime,
	}
}

// toNotificationItems 批量转换
func (s *NotificationServiceServer) toNotificationItems(notifications []domain.Notification) []*notificationv2.NotificationItem {
	items := make([]*notificationv2.NotificationItem, 0, len(notifications))
	for _, n := range notifications {
		items = append(items, s.toNotificationItem(n))
	}
	return items
}

// toDomainTemplate 将 proto Template 转换为 domain.Template
func (s *NotificationServiceServer) toDomainTemplate(tpl *notificationv2.Template) domain.Template {
	return domain.Template{
		Id:                    tpl.GetId(),
		TemplateId:            tpl.GetTemplateId(),
		Channel:               domain.Channel(tpl.GetChannel()),
		Name:                  tpl.GetName(),
		Content:               tpl.GetContent(),
		Description:           tpl.GetDescription(),
		Status:                domain.TemplateStatus(tpl.GetStatus()),
		SMSSign:               tpl.GetSmsSign(),
		SMSProviderTemplateId: tpl.GetSmsProviderTemplateId(),
	}
}

// toProtoTemplate 将 domain.Template 转换为 proto Template
func (s *NotificationServiceServer) toProtoTemplate(tpl domain.Template) *notificationv2.Template {
	return &notificationv2.Template{
		Id:                    tpl.Id,
		TemplateId:            tpl.TemplateId,
		Channel:               notificationv2.Channel(tpl.Channel),
		Name:                  tpl.Name,
		Content:               tpl.Content,
		Description:           tpl.Description,
		Status:                int32(tpl.Status),
		SmsSign:               tpl.SMSSign,
		SmsProviderTemplateId: tpl.SMSProviderTemplateId,
		Ctime:                 tpl.Ctime,
	}
}
