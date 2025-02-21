// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: tag/v1/tag.proto

package tagv1

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
	TagService_CreateTag_FullMethodName  = "/tag.v1.TagService/CreateTag"
	TagService_AttachTags_FullMethodName = "/tag.v1.TagService/AttachTags"
	TagService_GetTags_FullMethodName    = "/tag.v1.TagService/GetTags"
	TagService_GetBizTags_FullMethodName = "/tag.v1.TagService/GetBizTags"
)

// TagServiceClient is the client API for TagService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TagServiceClient interface {
	CreateTag(ctx context.Context, in *CreateTagRequest, opts ...grpc.CallOption) (*CreateTagResponse, error)
	// 覆盖式的 API
	// 也就是直接用新的 tag 全部覆盖掉已有的 tag
	AttachTags(ctx context.Context, in *AttachTagsRequest, opts ...grpc.CallOption) (*AttachTagsResponse, error)
	// 我们可以预期，一个用户的标签不会有很多，所以没有特别大的必要做成分页
	GetTags(ctx context.Context, in *GetTagsRequest, opts ...grpc.CallOption) (*GetTagsResponse, error)
	GetBizTags(ctx context.Context, in *GetBizTagsRequest, opts ...grpc.CallOption) (*GetBizTagsResponse, error)
}

type tagServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTagServiceClient(cc grpc.ClientConnInterface) TagServiceClient {
	return &tagServiceClient{cc}
}

func (c *tagServiceClient) CreateTag(ctx context.Context, in *CreateTagRequest, opts ...grpc.CallOption) (*CreateTagResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateTagResponse)
	err := c.cc.Invoke(ctx, TagService_CreateTag_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tagServiceClient) AttachTags(ctx context.Context, in *AttachTagsRequest, opts ...grpc.CallOption) (*AttachTagsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AttachTagsResponse)
	err := c.cc.Invoke(ctx, TagService_AttachTags_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tagServiceClient) GetTags(ctx context.Context, in *GetTagsRequest, opts ...grpc.CallOption) (*GetTagsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetTagsResponse)
	err := c.cc.Invoke(ctx, TagService_GetTags_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tagServiceClient) GetBizTags(ctx context.Context, in *GetBizTagsRequest, opts ...grpc.CallOption) (*GetBizTagsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetBizTagsResponse)
	err := c.cc.Invoke(ctx, TagService_GetBizTags_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TagServiceServer is the server API for TagService service.
// All implementations must embed UnimplementedTagServiceServer
// for forward compatibility.
type TagServiceServer interface {
	CreateTag(context.Context, *CreateTagRequest) (*CreateTagResponse, error)
	// 覆盖式的 API
	// 也就是直接用新的 tag 全部覆盖掉已有的 tag
	AttachTags(context.Context, *AttachTagsRequest) (*AttachTagsResponse, error)
	// 我们可以预期，一个用户的标签不会有很多，所以没有特别大的必要做成分页
	GetTags(context.Context, *GetTagsRequest) (*GetTagsResponse, error)
	GetBizTags(context.Context, *GetBizTagsRequest) (*GetBizTagsResponse, error)
	mustEmbedUnimplementedTagServiceServer()
}

// UnimplementedTagServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTagServiceServer struct{}

func (UnimplementedTagServiceServer) CreateTag(context.Context, *CreateTagRequest) (*CreateTagResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTag not implemented")
}
func (UnimplementedTagServiceServer) AttachTags(context.Context, *AttachTagsRequest) (*AttachTagsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AttachTags not implemented")
}
func (UnimplementedTagServiceServer) GetTags(context.Context, *GetTagsRequest) (*GetTagsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTags not implemented")
}
func (UnimplementedTagServiceServer) GetBizTags(context.Context, *GetBizTagsRequest) (*GetBizTagsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBizTags not implemented")
}
func (UnimplementedTagServiceServer) mustEmbedUnimplementedTagServiceServer() {}
func (UnimplementedTagServiceServer) testEmbeddedByValue()                    {}

// UnsafeTagServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TagServiceServer will
// result in compilation errors.
type UnsafeTagServiceServer interface {
	mustEmbedUnimplementedTagServiceServer()
}

func RegisterTagServiceServer(s grpc.ServiceRegistrar, srv TagServiceServer) {
	// If the following call pancis, it indicates UnimplementedTagServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TagService_ServiceDesc, srv)
}

func _TagService_CreateTag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateTagRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TagServiceServer).CreateTag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TagService_CreateTag_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TagServiceServer).CreateTag(ctx, req.(*CreateTagRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TagService_AttachTags_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AttachTagsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TagServiceServer).AttachTags(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TagService_AttachTags_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TagServiceServer).AttachTags(ctx, req.(*AttachTagsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TagService_GetTags_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTagsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TagServiceServer).GetTags(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TagService_GetTags_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TagServiceServer).GetTags(ctx, req.(*GetTagsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TagService_GetBizTags_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBizTagsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TagServiceServer).GetBizTags(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TagService_GetBizTags_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TagServiceServer).GetBizTags(ctx, req.(*GetBizTagsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TagService_ServiceDesc is the grpc.ServiceDesc for TagService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TagService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tag.v1.TagService",
	HandlerType: (*TagServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateTag",
			Handler:    _TagService_CreateTag_Handler,
		},
		{
			MethodName: "AttachTags",
			Handler:    _TagService_AttachTags_Handler,
		},
		{
			MethodName: "GetTags",
			Handler:    _TagService_GetTags_Handler,
		},
		{
			MethodName: "GetBizTags",
			Handler:    _TagService_GetBizTags_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tag/v1/tag.proto",
}
