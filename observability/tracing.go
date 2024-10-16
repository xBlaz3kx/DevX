package observability

import (
	"context"
	"time"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type Tracing struct {
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
}

func (t *Tracing) Tracer() trace.Tracer {
	return t.tracer
}

func (t *Tracing) TracerProvider() trace.TracerProvider {
	return t.tracerProvider
}

func (t *Tracing) Shutdown(ctx context.Context) error {
	return t.tracerProvider.Shutdown(ctx)
}

// createNewExporter creates a new OTel over GRPC exporter.
func createNewExporter(ctx context.Context, config TracingConfig) (*otlptrace.Exporter, error) {
	conn, err := connectToBackend(ctx, config.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection to collector")
	}

	// Set up a GRPC trace exporter
	return otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
}

// NewTracing creates a new tracing instance
func NewTracing(ctx context.Context, info ServiceInfo, config TracingConfig) (*Tracing, error) {
	// Create trace exporter
	exporter, err := createNewExporter(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create trace exporter")
	}

	// Create tracer provider
	bsp := sdktrace.NewBatchSpanProcessor(
		exporter,
		sdktrace.WithBatchTimeout(time.Second*5),
		sdktrace.WithMaxQueueSize(500),
	)

	// Create a resource with attributes
	res, err := resource.New(ctx,
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
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set the global provider
	otel.SetTracerProvider(otelpyroscope.NewTracerProvider(tracerProvider))

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		jaeger.Jaeger{},
		b3.New(),
	))

	tracer := otel.Tracer(info.Name)

	return &Tracing{
		tracerProvider: tracerProvider,
		tracer:         tracer,
	}, nil
}
