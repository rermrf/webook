package interceptors

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"net"
	"strings"
)

type Builder struct {
}

// PeerName 获取对端应用名称
func (b *Builder) PeerName(ctx context.Context) string {
	return b.grpcHeaderValue(ctx, "app")
}

func (b *Builder) PeerIP(ctx context.Context) string {
	// 如果在 ctx 里面传入。或者说在客户端里面设置了，就直接用它设置的
	// 有时候你经过了网关之类的东西，就需要客户端主动设置，防止后面拿到网关的 IP
	clientIP := b.grpcHeaderValue(ctx, "client-ip")
	if clientIP != "" {
		return clientIP
	}

	// 从 grpc 里取对端 ip
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	if pr.Addr == net.Addr(nil) {
		return ""
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	if len(addSlice) > 1 {
		return addSlice[0]
	}
	return ""
}

func (b *Builder) grpcHeaderValue(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return strings.Join(md.Get(key), ";")
}
