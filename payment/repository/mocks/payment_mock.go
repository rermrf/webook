// Code generated by MockGen. DO NOT EDIT.
// Source: ./validator.go
//
// Generated by this command:
//
//	mockgen -source=./validator.go -destination=mocks/payment_mock.go --package=repomocks PaymentRepository
//

// Package repomocks is a generated GoMock package.
package repomocks

import (
	context "context"
	reflect "reflect"
	time "time"
	domain "webook/payment/domain"

	gomock "go.uber.org/mock/gomock"
)

// MockPaymentRepository is a mock of PaymentRepository interface.
type MockPaymentRepository struct {
	ctrl     *gomock.Controller
	recorder *MockPaymentRepositoryMockRecorder
	isgomock struct{}
}

// MockPaymentRepositoryMockRecorder is the mock recorder for MockPaymentRepository.
type MockPaymentRepositoryMockRecorder struct {
	mock *MockPaymentRepository
}

// NewMockPaymentRepository creates a new mock instance.
func NewMockPaymentRepository(ctrl *gomock.Controller) *MockPaymentRepository {
	mock := &MockPaymentRepository{ctrl: ctrl}
	mock.recorder = &MockPaymentRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPaymentRepository) EXPECT() *MockPaymentRepositoryMockRecorder {
	return m.recorder
}

// AddPayMent mocks base method.
func (m *MockPaymentRepository) AddPayMent(ctx context.Context, pmt domain.Payment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddPayMent", ctx, pmt)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddPayMent indicates an expected call of AddPayMent.
func (mr *MockPaymentRepositoryMockRecorder) AddPayMent(ctx, pmt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPayMent", reflect.TypeOf((*MockPaymentRepository)(nil).AddPayMent), ctx, pmt)
}

// FindExpiredPayment mocks base method.
func (m *MockPaymentRepository) FindExpiredPayment(ctx context.Context, offiset, limit int, t time.Time) ([]domain.Payment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindExpiredPayment", ctx, offiset, limit, t)
	ret0, _ := ret[0].([]domain.Payment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindExpiredPayment indicates an expected call of FindExpiredPayment.
func (mr *MockPaymentRepositoryMockRecorder) FindExpiredPayment(ctx, offiset, limit, t any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindExpiredPayment", reflect.TypeOf((*MockPaymentRepository)(nil).FindExpiredPayment), ctx, offiset, limit, t)
}

// GetPayment mocks base method.
func (m *MockPaymentRepository) GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPayment", ctx, bizTradeNO)
	ret0, _ := ret[0].(domain.Payment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPayment indicates an expected call of GetPayment.
func (mr *MockPaymentRepositoryMockRecorder) GetPayment(ctx, bizTradeNO any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPayment", reflect.TypeOf((*MockPaymentRepository)(nil).GetPayment), ctx, bizTradeNO)
}

// UpdatePayMent mocks base method.
func (m *MockPaymentRepository) UpdatePayMent(ctx context.Context, pmt domain.Payment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePayMent", ctx, pmt)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePayMent indicates an expected call of UpdatePayMent.
func (mr *MockPaymentRepositoryMockRecorder) UpdatePayMent(ctx, pmt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePayMent", reflect.TypeOf((*MockPaymentRepository)(nil).UpdatePayMent), ctx, pmt)
}
