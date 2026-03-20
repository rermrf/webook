package service

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"webook/notification/domain"
	"webook/notification/repository"
)

// TemplateService 通知模板服务接口
type TemplateService interface {
	// Create 创建模板
	Create(ctx context.Context, tpl domain.Template) (int64, error)
	// Update 更新模板
	Update(ctx context.Context, tpl domain.Template) error
	// GetByTemplateId 根据模板ID和渠道获取模板
	GetByTemplateId(ctx context.Context, templateId string, channel domain.Channel) (domain.Template, error)
	// List 按渠道分页查询模板
	List(ctx context.Context, channel domain.Channel, offset, limit int) ([]domain.Template, error)
	// Render 渲染模板内容
	Render(ctx context.Context, templateId string, channel domain.Channel, params map[string]string) (string, error)
}

type templateService struct {
	repo repository.TemplateRepository
}

func NewTemplateService(repo repository.TemplateRepository) TemplateService {
	return &templateService{
		repo: repo,
	}
}

func (s *templateService) Create(ctx context.Context, tpl domain.Template) (int64, error) {
	return s.repo.Create(ctx, tpl)
}

func (s *templateService) Update(ctx context.Context, tpl domain.Template) error {
	return s.repo.Update(ctx, tpl)
}

func (s *templateService) GetByTemplateId(ctx context.Context, templateId string, channel domain.Channel) (domain.Template, error) {
	return s.repo.FindByTemplateIdAndChannel(ctx, templateId, channel)
}

func (s *templateService) List(ctx context.Context, channel domain.Channel, offset, limit int) ([]domain.Template, error) {
	return s.repo.FindByChannel(ctx, channel, offset, limit)
}

func (s *templateService) Render(ctx context.Context, templateId string, channel domain.Channel, params map[string]string) (string, error) {
	tpl, err := s.repo.FindByTemplateIdAndChannel(ctx, templateId, channel)
	if err != nil {
		return "", err
	}
	// 检查模板状态
	if tpl.Status == domain.TemplateStatusDisabled {
		return "", fmt.Errorf("模板 %s 已禁用", templateId)
	}
	// 使用 text/template 渲染模板内容
	t, err := template.New("n").Parse(tpl.Content)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("渲染模板失败: %w", err)
	}
	return buf.String(), nil
}
