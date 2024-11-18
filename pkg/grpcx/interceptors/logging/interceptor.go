package logging

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
	"time"
	"webook/pkg/grpcx/interceptors"
	"webook/pkg/logger"
)

type InterceptorBuilder struct {
	L logger.LoggerV1
	interceptors.Builder
}

func (i *InterceptorBuilder) BuildClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//start := time.Now()
		//var event = "normal"
		defer func() {
			//duration := time.Since(start)
			//
			//fields := []logger.Field{
			//	logger.Int64("cost", duration.Milliseconds()),
			//	logger.String("type", "unary"),
			//	logger.String("method", info.FullMethod),
			//	logger.String("event", event),
			//	// 这个部分需要你的客户端配合，
			//	// 你需要知道是哪一个业务调用过来的
			//	// 是哪个业务哪个结点过来的
			//	logger.String("peer", i.PeerName(ctx)),
			//	logger.String("peer_ip", i.PeerIP(ctx)),
			//}
			//if err != nil {
			//	st, _ := status.FromError(err)
			//	fields = append(fields, logger.String("code", st.Code().String()), logger.String("code_msg", st.Message()))
			//}
			//i.L.Debug("RPC请求", fields...)
		}()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (i *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		var event = "normal"
		defer func() {
			duration := time.Since(start)
			if rec := recover(); rec != nil {
				switch recType := rec.(type) {
				case error:
					err = recType
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
				err = status.New(codes.Internal, "panic, err "+err.Error()).Err()
			}
			fields := []logger.Field{
				logger.Int64("cost", duration.Milliseconds()),
				logger.String("type", "unary"),
				logger.String("method", info.FullMethod),
				logger.String("event", event),
				// 这个部分需要你的客户端配合，
				// 你需要知道是哪一个业务调用过来的
				// 是哪个业务哪个结点过来的
				logger.String("peer", i.PeerName(ctx)),
				logger.String("peer_ip", i.PeerIP(ctx)),
			}
			if err != nil {
				st, _ := status.FromError(err)
				fields = append(fields, logger.String("code", st.Code().String()), logger.String("code_msg", st.Message()))
			}
			i.L.Debug("RPC请求", fields...)
		}()
		resp, err = handler(ctx, req)

		return
	}
}
