package ioc

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"time"
)

func InitOTEL() func(ctx context.Context) {
	res, err := newResource("webook", "v0.0.1")
	if err != nil {
		panic(err)
	}

	prop := newPropagator()
	// 在客户端和服务端之间传递 tracing 的相关信息
	otel.SetTextMapPropagator(prop)

	// 初始化 trace provider
	// 这个 provider 就是用来打点的时候构建
	tp, err := newTraceProvider(res)
	if err != nil {
		panic(err)
	}
	otel.SetTracerProvider(tp)
	return func(ctx context.Context) {
		tp.Shutdown(ctx)
	}
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		))
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	traceExporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// 默认是5秒，这里设置为1秒以演示目的。
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}

// 动态控制是否需要监控
//type MyTracerProvider struct {
//	Enable      bool
//	nopProvider trace.TracerProvider
//	provider    trace.TracerProvider
//}
//
//func (m *MyTracerProvider) tracerProvider() {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (m *MyTracerProvider) Tracer(name string, options ...trace2.TracerOption) trace2.Tracer {
//	if m.Enable {
//		return m.nopProvider.Tracer(name, options...)
//	}
//	return m.provider.Tracer(name, options...)
//}
