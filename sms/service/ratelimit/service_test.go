package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/pkg/ratelimit"
	"webook/pkg/ratelimit/mocks"
	"webook/sms/service"
	"webook/sms/service/mocks"
)

func TestRateLimitSMSService_Send(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (service.Service, ratelimit.Limiter)
		wantErr error
	}{
		{
			name: "正常发送",
			mock: func(ctrl *gomock.Controller) (service.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				svc.EXPECT().Send(context.Background(), "123456", []string{}, "15011111111").Return(nil)
				return svc, limiter
			},
			wantErr: nil,
		},
		{
			name: "触发限流",
			mock: func(ctrl *gomock.Controller) (service.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				return svc, limiter
			},
			wantErr: errors.New("触发了限流"),
		},
		{
			name: "限流器异常",
			mock: func(ctrl *gomock.Controller) (service.Service, ratelimit.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, errors.New("限流器异常"))
				return svc, limiter
			},
			wantErr: fmt.Errorf("短信服务判断是否限流出现问题，%w", errors.New("限流器异常")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			limitSvc := NewRateLimitSMSService(tc.mock(ctrl))
			err := limitSvc.Send(context.Background(), "123456", []string{}, "15011111111")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
