package opentelemetry

import (
	"github.com/zhang-yong-feng/webz"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/zhang-yong-feng/webz"

// MiddlewareBuilder 应用OpenTelemetry与链路追踪的jeager和zipkin等技术结合
type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func (m *MiddlewareBuilder) Build() webz.HandleFunc {
	//先创建一个opentlrlmetry的trace实例
	if m.Tracer == nil {
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(ctx *webz.Context) {
		//获取上下文赋值给reqCtx
		reqCtx := ctx.Req.Context()
		reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Req.Header))
		reqCtx, span := m.Tracer.Start(reqCtx, "unKnown")
		defer span.End()
		span.SetAttributes(attribute.String("http.method", ctx.Req.Method))
		span.SetAttributes(attribute.String("http.url", ctx.Req.URL.String()))
		span.SetAttributes(attribute.String("http.scheme", ctx.Req.URL.Scheme))
		span.SetAttributes(attribute.String("http.host", ctx.Req.Host))
		ctx.Req = ctx.Req.WithContext(reqCtx)
		ctx.Next()
		span.SetName(ctx.FullPath)
	}

}
