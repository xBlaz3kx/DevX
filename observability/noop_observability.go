package observability

import (
	"context"
	"time"

	"github.com/GLCharge/otelzap"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func NewNoopObservability() Observability {
	return &noopObservability{}
}

type noopObservability struct {
}

func (n *noopObservability) Shutdown(ctx context.Context) {
}

func (n *noopObservability) Span(ctx context.Context, spanName string, fields ...zap.Field) (context.Context, func()) {
	return ctx, func() {}
}

func (n *noopObservability) LogSpan(ctx context.Context, spanName string, fields ...zap.Field) (context.Context, func(), otelzap.LoggerWithCtx) {
	return ctx, func() {}, otelzap.Ctx(ctx)
}

func (n *noopObservability) LogSpanWithTimeout(ctx context.Context, spanName string, timeout time.Duration, fields ...zap.Field) (context.Context, func(), otelzap.LoggerWithCtx) {
	return ctx, func() {}, otelzap.Ctx(ctx)
}

func (n *noopObservability) Log() *otelzap.Logger {
	return otelzap.L()
}

func (n *noopObservability) Metrics() *Metrics {
	return &Metrics{}
}

func (n *noopObservability) SetupHttpMiddleware(router *gin.Engine) {
}

func (n *noopObservability) WithSpanKind(spanKind trace.SpanKind) *Impl {
	return &Impl{}
}
