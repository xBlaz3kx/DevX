package observability

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/xBlaz3kx/DevX/tls"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MetricsConfig configures the metrics for the application (over OpenTelemetry GRPC).
type MetricsConfig struct {
	Enabled      bool    `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Address      string  `json:"address,omitempty" yaml:"address" mapstructure:"address"`
	TLS          tls.TLS `json:"tls" yaml:"tls" mapstructure:"tls"`
	PushInterval int64   `json:"pushInterval,omitempty" yaml:"pushInterval" mapstructure:"pushInterval"`
}

type Metrics struct {
	http   httpMetrics
	rabbit rabbitMetrics
}

// NewMetrics Creates a new metrics instance
func NewMetrics(ctx context.Context, info ServiceInfo, config MetricsConfig) (*Metrics, error) {
	options := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(config.Address),
	}

	if config.TLS.IsEnabled {
		options = append(options, otlpmetricgrpc.WithTLSCredentials(nil))
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

	httpMetrics, err := newHttpMetrics()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http metrics")
	}

	rabbitMetrics, err := newRabbitMetrics()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create rabbit metrics")
	}

	return &Metrics{
		http:   httpMetrics,
		rabbit: rabbitMetrics,
	}, nil
}

func connectToBackend(ctx context.Context, address string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	return grpc.NewClient(address, opts...)
}
