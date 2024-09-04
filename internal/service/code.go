package service

import (
	"context"
	"fmt"
	"math/rand"
	"webook/internal/repository"
	"webook/internal/service/sms"
)

const codeTplId = "tplId"

var (
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
)

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
	//tplId  string
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	// TODO: 生成一个验证码
	code := svc.generateCode()
	// 塞到 Redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	return svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
}

func (svc *CodeService) generateCode() string {
	// 生成 num 在 0， 999999 之间，包含 0 和 999999
	num := rand.Intn(1000000)
	// 不够六位的补前导0
	return fmt.Sprintf("%6d", num)
}

func (svc *CodeService) Verify(ctx context.Context, biz string, inputCode string, phone string) (bool, error) {
	return svc.repo.Verify(ctx, biz, inputCode, phone)
}
