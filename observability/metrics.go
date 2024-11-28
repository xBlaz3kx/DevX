package observability

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xBlaz3kx/DevX/tls"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	attrStatus  = "status"
	attrPath    = "path"
	attrMethod  = "method"
	attrUserId  = "user_id"
	attrTraceId = "trace_id"
	attrSpanId  = "span_id"
)

// MetricsConfig configures the metrics for the application (over OpenTelemetry GRPC).
type MetricsConfig struct {
	Enabled      bool    `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Address      string  `json:"address,omitempty" yaml:"address" mapstructure:"address"`
	TLS          tls.TLS `json:"tls" yaml:"tls" mapstructure:"tls"`
	PushInterval int64   `json:"pushInterval,omitempty" yaml:"pushInterval" mapstructure:"pushInterval"`
}

type Metrics struct {
	http httpMetrics
}

// NewMetrics Creates a new metrics instance
func NewMetrics(ctx context.Context, info ServiceInfo, config MetricsConfig) (*Metrics, error) {
	options := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(config.Address),
	}

	if config.TLS.IsEnabled {
		tlsConfig, err := config.TLS.ToTlsConfig()
		if err != nil {
			return nil, err
		}

		options = append(options, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		options = append(options, otlpmetricgrpc.WithInsecure())
	}

	conn, err := connectToBackend(ctx, config.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection to collector")
	}

	options = append(options, otlpmetricgrpc.WithGRPCConn(conn))

	exporter, err := otlpmetricgrpc.New(ctx, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create otlp metric exporter")
	}

	if config.PushInterval == 0 {
		config.PushInterval = 5
	}

	resource, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(info.Name),
			semconv.ServiceVersionKey.String(info.Version),
		),
		resource.WithFromEnv(),
		resource.WithContainer(),
		resource.WithOS(),
		resource.WithOSType(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create resource")
	}

	meterProvider := metricsdk.NewMeterProvider(
		metricsdk.WithReader(
			metricsdk.NewPeriodicReader(
				exporter,
				metricsdk.WithInterval(time.Duration(config.PushInterval)*time.Second),
			),
		),
		metricsdk.WithResource(resource),
	)

	otel.SetMeterProvider(meterProvider)

	httpMetrics, err := newHttpMetrics(info.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http metrics")
	}

	return &Metrics{
		http: httpMetrics,
	}, nil
}

type MetricMiddlewareOpts struct {
	tracingEnabled       bool
	userHeader           string
	additionalAttributes []attribute.KeyValue
}

func defaultMetricMiddlewareOpts() *MetricMiddlewareOpts {
	return &MetricMiddlewareOpts{
		tracingEnabled: false,
	}
}

type MetricMiddlewareOpt func(*MetricMiddlewareOpts)

func WithTracingEnabled(enabled bool) MetricMiddlewareOpt {
	return func(opts *MetricMiddlewareOpts) {
		opts.tracingEnabled = enabled
	}
}

func WithUserHeader(header string) MetricMiddlewareOpt {
	return func(opts *MetricMiddlewareOpts) {
		opts.userHeader = header
	}
}

func WithAdditionalAttributes(attrs []attribute.KeyValue) MetricMiddlewareOpt {
	return func(opts *MetricMiddlewareOpts) {
		opts.additionalAttributes = attrs
	}
}

// Middleware returns a gin middleware that records metrics for each HTTP request.
func (m *Metrics) Middleware(opts ...MetricMiddlewareOpt) func(c *gin.Context) {
	return func(c *gin.Context) {
		if m == nil {
			c.Next()
			return
		}

		// Defaults
		middlewareOpts := defaultMetricMiddlewareOpts()

		// Apply middleware opts
		for _, opt := range opts {
			opt(middlewareOpts)
		}

		ctx := c.Request.Context()
		startTime := time.Now()

		// Hardcoded attributes. These are always present in the metrics
		attrs := []attribute.KeyValue{
			attribute.String(attrPath, c.FullPath()),
			attribute.String(attrMethod, c.Request.Method),
		}

		// Add additional attributes to the metrics if they are present
		if len(middlewareOpts.additionalAttributes) > 0 {
			attrs = append(attrs, middlewareOpts.additionalAttributes...)
		}

		// Add user ID to the metrics if it is present in the request headers
		if middlewareOpts.userHeader != "" && c.GetHeader(middlewareOpts.userHeader) != "" {
			attrs = append(attrs, attribute.String(attrUserId, c.GetHeader(middlewareOpts.userHeader)))
		}

		// Form an exemplar for the metrics
		if middlewareOpts.tracingEnabled {
			span := trace.SpanFromContext(ctx)
			// Add trace ID and span ID to the metrics if the span is being recorded,
			// forming an exemplar
			if span.IsRecording() {
				attrs = append(
					attrs,
					attribute.String(attrTraceId, span.SpanContext().TraceID().String()),
					attribute.String(attrSpanId, span.SpanContext().SpanID().String()),
				)
			}
		}

		// Continue processing the request
		c.Next()

		// Append the writer status after the request has been processed
		attrs = append(attrs, attribute.Int(attrStatus, c.Writer.Status()))

		attributes := metric.WithAttributes(attrs...)

		// Record request duration
		m.http.requestDuration.Record(ctx, time.Since(startTime).Seconds(), attributes)

		// Record request count
		m.http.requestsTotal.Add(ctx, 1, attributes)
	}
}

func connectToBackend(ctx context.Context, address string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	return grpc.NewClient(address, opts...)
}
