package service

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"webook/internal/repository"
	"webook/internal/service/sms"
)

const codeTplId = "tplId"

var (
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
)

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, inputCode string, phone string) (bool, error)
}

type CodeServiceImpl struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
	//tplId  string
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &CodeServiceImpl{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *CodeServiceImpl) Send(ctx context.Context, biz string, phone string) error {
	// TODO: 生成一个验证码
	code := svc.generateCode()
	// 塞到 Redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	if err != nil {
		zap.L().Warn("发送太频繁", zap.Error(err))
	}
	return err
}

func (svc *CodeServiceImpl) generateCode() string {
	// 生成 num 在 0， 999999 之间，包含 0 和 999999
	num := rand.Intn(1000000)
	// 不够六位的补前导0
	return fmt.Sprintf("%06d", num)
}

func (svc *CodeServiceImpl) Verify(ctx context.Context, biz string, inputCode string, phone string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}
