package trace

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"webook/pkg/grpcx/interceptors"
)

type InterceptorBuilder struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	interceptors.Builder
}

func (b *InterceptorBuilder) BuildClient() grpc.UnaryClientInterceptor {
	propagator := b.propagator
	if propagator == nil {
		// 这个是全局
		propagator = otel.GetTextMapPropagator()
	}
	tracer := b.tracer
	if tracer == nil {
		tracer = otel.Tracer("webook/pkg/grpcx/interceptors/trace")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("server"),
	}
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attrs...))
		defer span.End()
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(e.Code)))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
		}()

		// inject 过程
		// 要把跟 trace 有关的链路元数据，传递到服务端
		ctx = inject(ctx, propagator)
		err = invoker(ctx, method, req, reply, cc, opts...)
		return err
	}
}

func (b *InterceptorBuilder) BuildServer() grpc.UnaryServerInterceptor {
	propagator := b.propagator
	if propagator == nil {
		// 这个是全局
		propagator = otel.GetTextMapPropagator()
	}
	tracer := b.tracer
	if tracer == nil {
		tracer = otel.Tracer("webook/pkg/grpcx/interceptors/trace")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("server"),
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = extract(ctx, propagator)
		ctx, span := tracer.Start(
			ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attrs...),
		)
		defer span.End()
		span.SetAttributes(
			semconv.RPCMethodKey.String(info.FullMethod),
			attribute.Key("net.peer.name").String(b.PeerName(ctx)),
			attribute.Key("net.peer.ip").String(b.PeerIP(ctx)),
		)
		defer func() {
			if err != nil {
				span.RecordError(err)
			} else {
				span.SetStatus(codes.Ok, "ok")
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func inject(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	// 先看 ctx 里面有没有元数据
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	// 把元数据放回去 ctx，具体怎么放，放什么内容，由 propagator 决定
	propagators.Inject(ctx, GRPCHeaderCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

func extract(ctx context.Context, p propagation.TextMapPropagator) context.Context {
	// 拿到客户端过来的链路元数据
	// "md": map[string]string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	// 把这个 md 注入到 ctx 中
	// 根据你采用 zipkin 或者 jeager，他的注入方式不同
	return p.Extract(ctx, GRPCHeaderCarrier(md))

}

type GRPCHeaderCarrier metadata.MD

func (g GRPCHeaderCarrier) Get(key string) string {
	vals := metadata.MD(g).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (g GRPCHeaderCarrier) Set(key string, value string) {
	metadata.MD(g).Set(key, value)
}

func (g GRPCHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(g))
	for k := range metadata.MD(g) {
		keys = append(keys, k)
	}
	return keys
}
