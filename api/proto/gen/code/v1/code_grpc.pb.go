// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: code/v1/code.proto

package codev1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	CodeService_Send_FullMethodName   = "/code.v1.CodeService/Send"
	CodeService_Verify_FullMethodName = "/code.v1.CodeService/Verify"
)

// CodeServiceClient is the client API for CodeService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CodeServiceClient interface {
	Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (*SendResponse, error)
	Verify(ctx context.Context, in *VerifyRequest, opts ...grpc.CallOption) (*VerifyResponse, error)
}

type codeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCodeServiceClient(cc grpc.ClientConnInterface) CodeServiceClient {
	return &codeServiceClient{cc}
}

func (c *codeServiceClient) Send(ctx context.Context, in *SendRequest, opts ...grpc.CallOption) (*SendResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendResponse)
	err := c.cc.Invoke(ctx, CodeService_Send_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *codeServiceClient) Verify(ctx context.Context, in *VerifyRequest, opts ...grpc.CallOption) (*VerifyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VerifyResponse)
	err := c.cc.Invoke(ctx, CodeService_Verify_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CodeServiceServer is the server API for CodeService service.
// All implementations must embed UnimplementedCodeServiceServer
// for forward compatibility.
type CodeServiceServer interface {
	Send(context.Context, *SendRequest) (*SendResponse, error)
	Verify(context.Context, *VerifyRequest) (*VerifyResponse, error)
	mustEmbedUnimplementedCodeServiceServer()
}

// UnimplementedCodeServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedCodeServiceServer struct{}

func (UnimplementedCodeServiceServer) Send(context.Context, *SendRequest) (*SendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Send not implemented")
}
func (UnimplementedCodeServiceServer) Verify(context.Context, *VerifyRequest) (*VerifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Verify not implemented")
}
func (UnimplementedCodeServiceServer) mustEmbedUnimplementedCodeServiceServer() {}
func (UnimplementedCodeServiceServer) testEmbeddedByValue()                     {}

// UnsafeCodeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CodeServiceServer will
// result in compilation errors.
type UnsafeCodeServiceServer interface {
	mustEmbedUnimplementedCodeServiceServer()
}

func RegisterCodeServiceServer(s grpc.ServiceRegistrar, srv CodeServiceServer) {
	// If the following call pancis, it indicates UnimplementedCodeServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CodeService_ServiceDesc, srv)
}

func _CodeService_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CodeServiceServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CodeService_Send_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CodeServiceServer).Send(ctx, req.(*SendRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CodeService_Verify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CodeServiceServer).Verify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CodeService_Verify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CodeServiceServer).Verify(ctx, req.(*VerifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CodeService_ServiceDesc is the grpc.ServiceDesc for CodeService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CodeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "code.v1.CodeService",
	HandlerType: (*CodeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _CodeService_Send_Handler,
		},
		{
			MethodName: "Verify",
			Handler:    _CodeService_Verify_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "code/v1/code.proto",
}
