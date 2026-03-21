package repository

import (
	"context"
	"encoding/json"
	"time"

	"webook/notification/domain"
	"webook/notification/repository/cache"
	"webook/notification/repository/dao"
	"webook/pkg/logger"
)

// NotificationRepository 通知仓储接口
type NotificationRepository interface {
	Create(ctx context.Context, n domain.Notification) (int64, error)
	BatchCreate(ctx context.Context, ns []domain.Notification) ([]int64, error)
	FindByKeyAndChannel(ctx context.Context, key string, channel domain.Channel) (domain.Notification, error)
	FindByUserId(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error)
	FindByUserIdAndGroup(ctx context.Context, userId int64, group domain.NotificationGroup, offset, limit int) ([]domain.Notification, error)
	FindUnreadByUserId(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error)
	MarkAsRead(ctx context.Context, userId int64, ids []int64) error
	MarkAllAsRead(ctx context.Context, userId int64) error
	GetUnreadCount(ctx context.Context, userId int64) (domain.UnreadCount, error)
	UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus) error
	Delete(ctx context.Context, userId int64, id int64) error
	DeleteAll(ctx context.Context, userId int64) error
	FindScheduledReady(ctx context.Context, limit int) ([]domain.Notification, error)
}

type CachedNotificationRepository struct {
	dao   dao.NotificationDAO
	cache cache.NotificationCache
	l     logger.LoggerV1
}

func NewCachedNotificationRepository(
	dao dao.NotificationDAO,
	cache cache.NotificationCache,
	l logger.LoggerV1,
) NotificationRepository {
	return &CachedNotificationRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (r *CachedNotificationRepository) Create(ctx context.Context, n domain.Notification) (int64, error) {
	entity := r.toEntity(n)
	id, err := r.dao.Insert(ctx, entity)
	if err != nil {
		return 0, err
	}
	// 站内通知才更新未读缓存和推送SSE
	if n.Channel == domain.ChannelInApp {
		er := r.cache.IncrUnreadCount(ctx, n.UserId, uint8(n.GroupType))
		if er != nil {
			r.l.Error("增加未读数缓存失败",
				logger.Int64("userId", n.UserId),
				logger.Error(er))
		}
		go func() {
			r.publishSSE(n.UserId, n)
		}()
	}
	return id, nil
}

func (r *CachedNotificationRepository) BatchCreate(ctx context.Context, ns []domain.Notification) ([]int64, error) {
	if len(ns) == 0 {
		return nil, nil
	}
	entities := make([]dao.Notification, 0, len(ns))
	for _, n := range ns {
		entities = append(entities, r.toEntity(n))
	}
	ids, err := r.dao.BatchInsert(ctx, entities)
	if err != nil {
		return nil, err
	}
	// 批量更新缓存并推送SSE
	for _, n := range ns {
		if n.Channel == domain.ChannelInApp {
			er := r.cache.IncrUnreadCount(ctx, n.UserId, uint8(n.GroupType))
			if er != nil {
				r.l.Error("增加未读数缓存失败",
					logger.Int64("userId", n.UserId),
					logger.Error(er))
			}
			go func(notification domain.Notification) {
				r.publishSSE(notification.UserId, notification)
			}(n)
		}
	}
	return ids, nil
}

func (r *CachedNotificationRepository) FindByKeyAndChannel(ctx context.Context, key string, channel domain.Channel) (domain.Notification, error) {
	entity, err := r.dao.FindByKeyAndChannel(ctx, key, uint8(channel))
	if err != nil {
		return domain.Notification{}, err
	}
	return r.toDomain(entity), nil
}

func (r *CachedNotificationRepository) FindByUserId(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error) {
	entities, err := r.dao.FindByUserId(ctx, userId, offset, limit)
	if err != nil {
		return nil, err
	}
	return r.toDomainList(entities), nil
}

func (r *CachedNotificationRepository) FindByUserIdAndGroup(ctx context.Context, userId int64, group domain.NotificationGroup, offset, limit int) ([]domain.Notification, error) {
	entities, err := r.dao.FindByUserIdAndGroup(ctx, userId, uint8(group), offset, limit)
	if err != nil {
		return nil, err
	}
	return r.toDomainList(entities), nil
}

func (r *CachedNotificationRepository) FindUnreadByUserId(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error) {
	entities, err := r.dao.FindUnreadByUserId(ctx, userId, offset, limit)
	if err != nil {
		return nil, err
	}
	return r.toDomainList(entities), nil
}

func (r *CachedNotificationRepository) MarkAsRead(ctx context.Context, userId int64, ids []int64) error {
	err := r.dao.MarkAsRead(ctx, userId, ids)
	if err != nil {
		return err
	}
	_ = r.cache.ClearUnreadCount(ctx, userId)
	return nil
}

func (r *CachedNotificationRepository) MarkAllAsRead(ctx context.Context, userId int64) error {
	err := r.dao.MarkAllAsRead(ctx, userId)
	if err != nil {
		return err
	}
	return r.cache.ClearUnreadCount(ctx, userId)
}

func (r *CachedNotificationRepository) GetUnreadCount(ctx context.Context, userId int64) (domain.UnreadCount, error) {
	// 先从缓存获取
	byGroup, total, err := r.cache.GetUnreadCount(ctx, userId)
	if err == nil {
		result := domain.UnreadCount{
			Total:   total,
			ByGroup: make(map[domain.NotificationGroup]int64),
		}
		for g, c := range byGroup {
			result.ByGroup[domain.NotificationGroup(g)] = c
		}
		return result, nil
	}
	// 缓存不存在，从数据库加载
	countByGroup, err := r.dao.CountUnreadByGroup(ctx, userId)
	if err != nil {
		return domain.UnreadCount{}, err
	}
	// 转换并回写缓存
	result := domain.UnreadCount{
		ByGroup: make(map[domain.NotificationGroup]int64),
	}
	cacheData := make(map[uint8]int64)
	for g, c := range countByGroup {
		result.ByGroup[domain.NotificationGroup(g)] = c
		result.Total += c
		cacheData[g] = c
	}
	go func() {
		er := r.cache.SetUnreadCount(context.Background(), userId, cacheData)
		if er != nil {
			r.l.Error("回写未读数缓存失败",
				logger.Int64("userId", userId),
				logger.Error(er))
		}
	}()
	return result, nil
}

func (r *CachedNotificationRepository) UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus) error {
	return r.dao.UpdateStatus(ctx, id, uint8(status))
}

func (r *CachedNotificationRepository) Delete(ctx context.Context, userId int64, id int64) error {
	err := r.dao.Delete(ctx, userId, id)
	if err != nil {
		return err
	}
	_ = r.cache.ClearUnreadCount(ctx, userId)
	return nil
}

func (r *CachedNotificationRepository) DeleteAll(ctx context.Context, userId int64) error {
	err := r.dao.DeleteByUserId(ctx, userId)
	if err != nil {
		return err
	}
	return r.cache.ClearUnreadCount(ctx, userId)
}

func (r *CachedNotificationRepository) publishSSE(userId int64, n domain.Notification) {
	// 获取当前未读数
	byGroup, total, _ := r.cache.GetUnreadCount(context.Background(), userId)

	msg := map[string]any{
		"type":     "notification",
		"total":    total,
		"by_group": byGroup,
		"notification": map[string]any{
			"id":           n.Id,
			"group_type":   int(n.GroupType),
			"source_id":    n.SourceId,
			"source_name":  n.SourceName,
			"target_id":    n.TargetId,
			"target_type":  n.TargetType,
			"target_title": n.TargetTitle,
			"content":      n.Content,
			"ctime":        n.Ctime,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		r.l.Error("序列化SSE消息失败", logger.Error(err))
		return
	}

	payload, err := json.Marshal(map[string]any{
		"user_id": userId,
		"data":    json.RawMessage(data),
	})
	if err != nil {
		r.l.Error("序列化SSE payload失败", logger.Error(err))
		return
	}

	er := r.cache.PublishSSE(context.Background(), userId, payload)
	if er != nil {
		r.l.Error("推送SSE通知失败",
			logger.Int64("userId", userId),
			logger.Error(er))
	}
}

func (r *CachedNotificationRepository) FindScheduledReady(ctx context.Context, limit int) ([]domain.Notification, error) {
	now := time.Now().UnixMilli()
	entities, err := r.dao.FindScheduledReady(ctx, now, limit)
	if err != nil {
		return nil, err
	}
	return r.toDomainList(entities), nil
}

func (r *CachedNotificationRepository) toEntity(n domain.Notification) dao.Notification {
	params := ""
	if len(n.TemplateParams) > 0 {
		data, _ := json.Marshal(n.TemplateParams)
		params = string(data)
	}
	return dao.Notification{
		KeyField:       n.Key,
		BizId:          n.BizId,
		Channel:        uint8(n.Channel),
		Receiver:       n.Receiver,
		UserId:         n.UserId,
		TemplateId:     n.TemplateId,
		TemplateParams: params,
		Content:        n.Content,
		Status:         uint8(n.Status),
		Strategy:       uint8(n.Strategy),
		ScheduledTime:  n.ScheduledTime,
		GroupType:      uint8(n.GroupType),
		SourceId:       n.SourceId,
		SourceName:     n.SourceName,
		TargetId:       n.TargetId,
		TargetType:     n.TargetType,
		TargetTitle:    n.TargetTitle,
		IsRead:         n.IsRead,
	}
}

func (r *CachedNotificationRepository) toDomain(e dao.Notification) domain.Notification {
	var params map[string]string
	if e.TemplateParams != "" {
		_ = json.Unmarshal([]byte(e.TemplateParams), &params)
	}
	return domain.Notification{
		Id:             e.Id,
		Key:            e.KeyField,
		BizId:          e.BizId,
		Channel:        domain.Channel(e.Channel),
		Receiver:       e.Receiver,
		UserId:         e.UserId,
		TemplateId:     e.TemplateId,
		TemplateParams: params,
		Content:        e.Content,
		Status:         domain.NotificationStatus(e.Status),
		Strategy:       domain.SendStrategy(e.Strategy),
		ScheduledTime:  e.ScheduledTime,
		GroupType:      domain.NotificationGroup(e.GroupType),
		SourceId:       e.SourceId,
		SourceName:     e.SourceName,
		TargetId:       e.TargetId,
		TargetType:     e.TargetType,
		TargetTitle:    e.TargetTitle,
		IsRead:         e.IsRead,
		Ctime:          e.Ctime,
		Utime:          e.Utime,
	}
}

func (r *CachedNotificationRepository) toDomainList(entities []dao.Notification) []domain.Notification {
	result := make([]domain.Notification, 0, len(entities))
	for _, e := range entities {
		result = append(result, r.toDomain(e))
	}
	return result
}
