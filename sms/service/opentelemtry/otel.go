package opentelemtry

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"webook/sms/service"
)

type TracingOTELService struct {
	svc    service.Service
	tracer trace.Tracer
}

func NewTracingOTELService(svc service.Service) *TracingOTELService {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("/webook/sms/service/sms/opentelemtry")
	return &TracingOTELService{
		svc:    svc,
		tracer: tracer,
	}
}

func (s *TracingOTELService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	ctx, span := s.tracer.Start(ctx,
		"sms"+biz,
		// 因为我是一个调用短信服务的客户端
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End(trace.WithStackTrace(true))

	err := s.svc.Send(ctx, biz, args)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
