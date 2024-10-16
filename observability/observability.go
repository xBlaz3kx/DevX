package observability

import (
	"context"
	"time"

	"github.com/GLCharge/otelzap"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/uptrace/opentelemetry-go-extra/otelutil"
	"github.com/xBlaz3kx/DevX/util"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Impl struct {
	config Config
	info   ServiceInfo

	logging   *Logging
	tracing   *Tracing
	profiling *Profiling
	metrics   *Metrics
	metric.MeterProvider

	spanKind trace.SpanKind
}

type Observability interface {
	Shutdown(ctx context.Context) error
	Span(ctx context.Context, spanName string, fields ...zap.Field) (context.Context, func())
	LogSpan(ctx context.Context, spanName string, fields ...zap.Field) (context.Context, func(), otelzap.LoggerWithCtx)
	LogSpanWithTimeout(ctx context.Context, spanName string, timeout time.Duration, fields ...zap.Field) (context.Context, func(), otelzap.LoggerWithCtx)
	Log() *otelzap.Logger
	Metrics() *Metrics
	SetupGinMiddleware(router *gin.Engine)
	WithSpanKind(spanKind trace.SpanKind) *Impl
	metric.MeterProvider
}

func NewObservability(ctx context.Context, info ServiceInfo, config Config) (*Impl, error) {
	obs := Impl{
		config:   config,
		info:     info,
		logging:  NewLogging(config.Logging),
		spanKind: trace.SpanKindInternal,
	}

	var err error

	// Setup tracing
	if config.Tracing.Enabled {
		obs.tracing, err = NewTracing(ctx, info, config.Tracing)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup tracing")
		}
		obs.Log().Info("Tracing enabled")
	}

	// Setup profiling
	if config.Profiling.Enabled {
		obs.profiling, err = NewProfiler(info.Name, config.Profiling)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup profiling")
		}
		obs.Log().Info("Profiling enabled")
	}

	// Setup metrics
	if config.Metrics.Enabled {
		// General metrics
		obs.metrics, err = NewMetrics(ctx, info, config.Metrics)
		if err != nil {
			return nil, errors.Wrap(err, "failed to setup metrics")
		}
		obs.Log().Info("Metrics enabled")

		obs.MeterProvider = otel.GetMeterProvider()
	}

	return &obs, nil
}

// Shutdown Gracefully shutdown observability components
func (obs *Impl) Shutdown(ctx context.Context) error {
	if !util.IsNilInterfaceOrPointer(obs.tracing) {
		err := obs.tracing.Shutdown(ctx)
		if err != nil {
			return err
		}
	}

	if !util.IsNilInterfaceOrPointer(obs.profiling) {
		err := obs.profiling.Shutdown()
		if err != nil {
			return err
		}
	}

	return obs.logging.Shutdown()
}

// Span creates a new span
func (obs *Impl) Span(ctx context.Context, spanName string, fields ...zap.Field) (context.Context, func()) {
	if obs == nil {
		return ctx, func() {}
	}

	endSpan := func() {}

	if obs.config.Tracing.Enabled {
		attrs := otelAttributesFromZapFields(fields)

		var span trace.Span
		ctx, span = obs.tracing.Tracer().Start(ctx, spanName, trace.WithAttributes(attrs...), trace.WithSpanKind(obs.spanKind))
		endSpan = func() {
			span.End()
		}
	}

	return ctx, endSpan
}

func otelAttributesFromZapFields(fields []zap.Field) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, len(fields))

	enc := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(enc)
	}

	for k, v := range enc.Fields {
		attrs = append(attrs, otelutil.Attribute(k, v))
	}
	return attrs
}

// LogSpan creates a new span with the given name and fields and logs the span and trace id fields
func (obs *Impl) LogSpan(ctx context.Context, spanName string, fields ...zap.Field) (context.Context, func(), otelzap.LoggerWithCtx) {
	if obs == nil {
		return ctx, func() {}, otelzap.LoggerWithCtx{}
	}

	ctx, end := obs.Span(ctx, spanName, fields...)

	logger := obs.logging.Logger().Ctx(ctx).With(fields...)

	return ctx, end, logger
}

// LogSpanWithTimeout creates a new span with the given name and fields and logs the span and trace id fields
func (obs *Impl) LogSpanWithTimeout(ctx context.Context, spanName string, timeout time.Duration, fields ...zap.Field) (context.Context, func(), otelzap.LoggerWithCtx) {
	ctx, end := obs.Span(ctx, spanName, fields...)
	ctx, cancel := context.WithTimeout(ctx, timeout)

	logger := obs.logging.Logger().Ctx(ctx).With(fields...)

	return ctx, func() {
		end()
		cancel()
	}, logger
}

// Log Returns the logger
func (obs *Impl) Log() *otelzap.Logger {
	if obs == nil {
		return otelzap.L()
	}

	return obs.logging.Logger()
}

// Metrics returns the metrics instance
func (obs *Impl) Metrics() *Metrics {
	if obs == nil {
		return nil
	}

	return obs.metrics
}

// SetupGinMiddleware adds middleware to the Gin router based on the observability configuration
func (obs *Impl) SetupGinMiddleware(router *gin.Engine) {
	if obs == nil {
		return
	}

	obsServer := obs.clone()
	obsServer.spanKind = trace.SpanKindServer

	if obsServer.config.Tracing.Enabled {
		router.Use(otelgin.Middleware(obsServer.info.Name))
	}

	if obsServer.config.Metrics.Enabled {
		router.Use(obsServer.Metrics().Middleware())
	}
}

// Clone returns a copy of the observability instance
func (obs *Impl) clone() *Impl {
	return &Impl{
		config: obs.config,
		info:   obs.info,

		logging: obs.logging,
		tracing: obs.tracing,
		metrics: obs.metrics,

		spanKind: obs.spanKind,
	}
}

// WithSpanKind returns a copy of the observability instance with the given span type
func (obs *Impl) WithSpanKind(spanKind trace.SpanKind) *Impl {
	o := obs.clone()
	o.spanKind = spanKind
	return o
}
