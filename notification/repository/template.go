package repository

import (
	"context"
	"encoding/json"

	"webook/notification/domain"
	"webook/notification/repository/cache"
	"webook/notification/repository/dao"
	"webook/pkg/logger"
)

// TemplateRepository 通知模板仓储接口
type TemplateRepository interface {
	Create(ctx context.Context, t domain.Template) (int64, error)
	Update(ctx context.Context, t domain.Template) error
	FindByTemplateIdAndChannel(ctx context.Context, templateId string, channel domain.Channel) (domain.Template, error)
	FindByChannel(ctx context.Context, channel domain.Channel, offset, limit int) ([]domain.Template, error)
}

type CachedTemplateRepository struct {
	dao   dao.TemplateDAO
	cache cache.TemplateCache
	l     logger.LoggerV1
}

func NewCachedTemplateRepository(
	dao dao.TemplateDAO,
	cache cache.TemplateCache,
	l logger.LoggerV1,
) TemplateRepository {
	return &CachedTemplateRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (r *CachedTemplateRepository) Create(ctx context.Context, t domain.Template) (int64, error) {
	return r.dao.Insert(ctx, r.toEntity(t))
}

func (r *CachedTemplateRepository) Update(ctx context.Context, t domain.Template) error {
	err := r.dao.Update(ctx, r.toEntity(t))
	if err != nil {
		return err
	}
	// 删除缓存
	er := r.cache.Del(ctx, t.TemplateId, uint8(t.Channel))
	if er != nil {
		r.l.Error("删除模板缓存失败",
			logger.String("templateId", t.TemplateId),
			logger.Error(er))
	}
	return nil
}

func (r *CachedTemplateRepository) FindByTemplateIdAndChannel(ctx context.Context, templateId string, channel domain.Channel) (domain.Template, error) {
	// 先从缓存获取
	data, err := r.cache.Get(ctx, templateId, uint8(channel))
	if err == nil {
		var t domain.Template
		if er := json.Unmarshal(data, &t); er == nil {
			return t, nil
		}
	}
	// 缓存未命中，从数据库加载
	entity, err := r.dao.FindByTemplateIdAndChannel(ctx, templateId, uint8(channel))
	if err != nil {
		return domain.Template{}, err
	}
	t := r.toDomain(entity)
	// 回写缓存
	go func() {
		data, er := json.Marshal(t)
		if er != nil {
			return
		}
		er = r.cache.Set(context.Background(), templateId, uint8(channel), data)
		if er != nil {
			r.l.Error("回写模板缓存失败",
				logger.String("templateId", templateId),
				logger.Error(er))
		}
	}()
	return t, nil
}

func (r *CachedTemplateRepository) FindByChannel(ctx context.Context, channel domain.Channel, offset, limit int) ([]domain.Template, error) {
	entities, err := r.dao.FindByChannel(ctx, uint8(channel), offset, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Template, 0, len(entities))
	for _, e := range entities {
		result = append(result, r.toDomain(e))
	}
	return result, nil
}

func (r *CachedTemplateRepository) toEntity(t domain.Template) dao.NotificationTemplate {
	return dao.NotificationTemplate{
		Id:                    t.Id,
		TemplateId:            t.TemplateId,
		Channel:               uint8(t.Channel),
		Name:                  t.Name,
		Content:               t.Content,
		Description:           t.Description,
		Status:                uint8(t.Status),
		SMSSign:               t.SMSSign,
		SMSProviderTemplateId: t.SMSProviderTemplateId,
	}
}

func (r *CachedTemplateRepository) toDomain(e dao.NotificationTemplate) domain.Template {
	return domain.Template{
		Id:                    e.Id,
		TemplateId:            e.TemplateId,
		Channel:               domain.Channel(e.Channel),
		Name:                  e.Name,
		Content:               e.Content,
		Description:           e.Description,
		Status:                domain.TemplateStatus(e.Status),
		SMSSign:               e.SMSSign,
		SMSProviderTemplateId: e.SMSProviderTemplateId,
		Ctime:                 e.Ctime,
		Utime:                 e.Utime,
	}
}
