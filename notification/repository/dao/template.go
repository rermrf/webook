package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// TemplateDAO 通知模板数据访问接口
type TemplateDAO interface {
	Insert(ctx context.Context, t NotificationTemplate) (int64, error)
	Update(ctx context.Context, t NotificationTemplate) error
	FindByTemplateIdAndChannel(ctx context.Context, templateId string, channel uint8) (NotificationTemplate, error)
	FindByChannel(ctx context.Context, channel uint8, offset, limit int) ([]NotificationTemplate, error)
}

type GORMTemplateDAO struct {
	db *gorm.DB
}

func NewGORMTemplateDAO(db *gorm.DB) TemplateDAO {
	return &GORMTemplateDAO{db: db}
}

func (d *GORMTemplateDAO) Insert(ctx context.Context, t NotificationTemplate) (int64, error) {
	now := time.Now().UnixMilli()
	t.Ctime = now
	t.Utime = now
	err := d.db.WithContext(ctx).Create(&t).Error
	if err != nil {
		return 0, err
	}
	return t.Id, nil
}

func (d *GORMTemplateDAO) Update(ctx context.Context, t NotificationTemplate) error {
	return d.db.WithContext(ctx).
		Model(&NotificationTemplate{}).
		Where("template_id = ? AND channel = ?", t.TemplateId, t.Channel).
		Updates(map[string]any{
			"name":                     t.Name,
			"content":                  t.Content,
			"description":              t.Description,
			"status":                   t.Status,
			"sms_sign":                 t.SMSSign,
			"sms_provider_template_id": t.SMSProviderTemplateId,
			"utime":                    time.Now().UnixMilli(),
		}).Error
}

func (d *GORMTemplateDAO) FindByTemplateIdAndChannel(ctx context.Context, templateId string, channel uint8) (NotificationTemplate, error) {
	var res NotificationTemplate
	err := d.db.WithContext(ctx).
		Where("template_id = ? AND channel = ?", templateId, channel).
		First(&res).Error
	return res, err
}

func (d *GORMTemplateDAO) FindByChannel(ctx context.Context, channel uint8, offset, limit int) ([]NotificationTemplate, error) {
	var res []NotificationTemplate
	err := d.db.WithContext(ctx).
		Where("channel = ? AND status = ?", channel, uint8(1)).
		Order("ctime DESC").
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

// NotificationTemplate 通知模板表
// 索引设计：
// 1. uk_template_channel: (template_id, channel) - 唯一索引
// 2. idx_channel_status: (channel, status) - 支持按渠道查询启用模板
type NotificationTemplate struct {
	Id                    int64  `gorm:"primaryKey;autoIncrement"`
	TemplateId            string `gorm:"type:varchar(128);uniqueIndex:uk_template_channel"`
	Channel               uint8  `gorm:"uniqueIndex:uk_template_channel;index:idx_channel_status"`
	Name                  string `gorm:"type:varchar(256)"`
	Content               string `gorm:"type:text;not null"`
	Description           string `gorm:"type:varchar(512)"`
	Status                uint8  `gorm:"index:idx_channel_status;default:1"`
	SMSSign               string `gorm:"column:sms_sign;type:varchar(64)"`
	SMSProviderTemplateId string `gorm:"column:sms_provider_template_id;type:varchar(128)"`
	Ctime                 int64
	Utime                 int64
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}
