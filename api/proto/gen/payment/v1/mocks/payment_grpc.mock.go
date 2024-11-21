// Code generated by MockGen. DO NOT EDIT.
// Source: ./api/proto/gen/payment/v1/payment_grpc.pb.go
//
// Generated by this command:
//
//	mockgen -source=./api/proto/gen/payment/v1/payment_grpc.pb.go -package=pmtmocks -destination=./api/proto/gen/payment/v1/mocks/payment_grpc.mock.go
//

// Package pmtmocks is a generated GoMock package.
package pmtmocks

import (
	context "context"
	reflect "reflect"
	pmtv1 "webook/api/proto/gen/payment/v1"

	gomock "go.uber.org/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockWechatPaymentServiceClient is a mock of WechatPaymentServiceClient interface.
type MockWechatPaymentServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockWechatPaymentServiceClientMockRecorder
}

// MockWechatPaymentServiceClientMockRecorder is the mock recorder for MockWechatPaymentServiceClient.
type MockWechatPaymentServiceClientMockRecorder struct {
	mock *MockWechatPaymentServiceClient
}

// NewMockWechatPaymentServiceClient creates a new mock instance.
func NewMockWechatPaymentServiceClient(ctrl *gomock.Controller) *MockWechatPaymentServiceClient {
	mock := &MockWechatPaymentServiceClient{ctrl: ctrl}
	mock.recorder = &MockWechatPaymentServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWechatPaymentServiceClient) EXPECT() *MockWechatPaymentServiceClientMockRecorder {
	return m.recorder
}

// GetPayment mocks base method.
func (m *MockWechatPaymentServiceClient) GetPayment(ctx context.Context, in *pmtv1.GetPaymentRequest, opts ...grpc.CallOption) (*pmtv1.GetPaymentResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetPayment", varargs...)
	ret0, _ := ret[0].(*pmtv1.GetPaymentResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPayment indicates an expected call of GetPayment.
func (mr *MockWechatPaymentServiceClientMockRecorder) GetPayment(ctx, in any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPayment", reflect.TypeOf((*MockWechatPaymentServiceClient)(nil).GetPayment), varargs...)
}

// NativePrePay mocks base method.
func (m *MockWechatPaymentServiceClient) NativePrePay(ctx context.Context, in *pmtv1.PrePayRequest, opts ...grpc.CallOption) (*pmtv1.NativePrePayResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "NativePrePay", varargs...)
	ret0, _ := ret[0].(*pmtv1.NativePrePayResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NativePrePay indicates an expected call of NativePrePay.
func (mr *MockWechatPaymentServiceClientMockRecorder) NativePrePay(ctx, in any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NativePrePay", reflect.TypeOf((*MockWechatPaymentServiceClient)(nil).NativePrePay), varargs...)
}

// MockWechatPaymentServiceServer is a mock of WechatPaymentServiceServer interface.
type MockWechatPaymentServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockWechatPaymentServiceServerMockRecorder
}

// MockWechatPaymentServiceServerMockRecorder is the mock recorder for MockWechatPaymentServiceServer.
type MockWechatPaymentServiceServerMockRecorder struct {
	mock *MockWechatPaymentServiceServer
}

// NewMockWechatPaymentServiceServer creates a new mock instance.
func NewMockWechatPaymentServiceServer(ctrl *gomock.Controller) *MockWechatPaymentServiceServer {
	mock := &MockWechatPaymentServiceServer{ctrl: ctrl}
	mock.recorder = &MockWechatPaymentServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWechatPaymentServiceServer) EXPECT() *MockWechatPaymentServiceServerMockRecorder {
	return m.recorder
}

// GetPayment mocks base method.
func (m *MockWechatPaymentServiceServer) GetPayment(arg0 context.Context, arg1 *pmtv1.GetPaymentRequest) (*pmtv1.GetPaymentResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPayment", arg0, arg1)
	ret0, _ := ret[0].(*pmtv1.GetPaymentResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPayment indicates an expected call of GetPayment.
func (mr *MockWechatPaymentServiceServerMockRecorder) GetPayment(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPayment", reflect.TypeOf((*MockWechatPaymentServiceServer)(nil).GetPayment), arg0, arg1)
}

// NativePrePay mocks base method.
func (m *MockWechatPaymentServiceServer) NativePrePay(arg0 context.Context, arg1 *pmtv1.PrePayRequest) (*pmtv1.NativePrePayResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NativePrePay", arg0, arg1)
	ret0, _ := ret[0].(*pmtv1.NativePrePayResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NativePrePay indicates an expected call of NativePrePay.
func (mr *MockWechatPaymentServiceServerMockRecorder) NativePrePay(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NativePrePay", reflect.TypeOf((*MockWechatPaymentServiceServer)(nil).NativePrePay), arg0, arg1)
}

// mustEmbedUnimplementedWechatPaymentServiceServer mocks base method.
func (m *MockWechatPaymentServiceServer) mustEmbedUnimplementedWechatPaymentServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedWechatPaymentServiceServer")
}

// mustEmbedUnimplementedWechatPaymentServiceServer indicates an expected call of mustEmbedUnimplementedWechatPaymentServiceServer.
func (mr *MockWechatPaymentServiceServerMockRecorder) mustEmbedUnimplementedWechatPaymentServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedWechatPaymentServiceServer", reflect.TypeOf((*MockWechatPaymentServiceServer)(nil).mustEmbedUnimplementedWechatPaymentServiceServer))
}

// MockUnsafeWechatPaymentServiceServer is a mock of UnsafeWechatPaymentServiceServer interface.
type MockUnsafeWechatPaymentServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeWechatPaymentServiceServerMockRecorder
}

// MockUnsafeWechatPaymentServiceServerMockRecorder is the mock recorder for MockUnsafeWechatPaymentServiceServer.
type MockUnsafeWechatPaymentServiceServerMockRecorder struct {
	mock *MockUnsafeWechatPaymentServiceServer
}

// NewMockUnsafeWechatPaymentServiceServer creates a new mock instance.
func NewMockUnsafeWechatPaymentServiceServer(ctrl *gomock.Controller) *MockUnsafeWechatPaymentServiceServer {
	mock := &MockUnsafeWechatPaymentServiceServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeWechatPaymentServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnsafeWechatPaymentServiceServer) EXPECT() *MockUnsafeWechatPaymentServiceServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedWechatPaymentServiceServer mocks base method.
func (m *MockUnsafeWechatPaymentServiceServer) mustEmbedUnimplementedWechatPaymentServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedWechatPaymentServiceServer")
}

// mustEmbedUnimplementedWechatPaymentServiceServer indicates an expected call of mustEmbedUnimplementedWechatPaymentServiceServer.
func (mr *MockUnsafeWechatPaymentServiceServerMockRecorder) mustEmbedUnimplementedWechatPaymentServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedWechatPaymentServiceServer", reflect.TypeOf((*MockUnsafeWechatPaymentServiceServer)(nil).mustEmbedUnimplementedWechatPaymentServiceServer))
}
