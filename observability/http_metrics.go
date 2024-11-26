package observability

import (
	"fmt"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type httpMetrics struct {
	requestsTotal   metric.Int64Counter
	requestDuration metric.Float64Histogram
}

const (
	httpRequestsTotal          = "http_requests_total"
	httpRequestDurationSeconds = "http_requests_duration_seconds"
	httpErrorsTotal            = "http_errors_total"
)

// Returns the metric name with the prefix
func getMetricsPrefix(prefix string, metric string) string {
	if prefix == "" {
		return metric
	}

	return fmt.Sprintf("%s_%s", prefix, metric)
}

// Initializes http meters
func newHttpMetrics(prefix string) (metrics httpMetrics, err error) {
	meter := otel.Meter("http")

	if metrics.requestsTotal, err = meter.Int64Counter(
		getMetricsPrefix(prefix, httpRequestsTotal),
		metric.WithDescription("Total number of HTTP requests"),
	); err != nil {
		return httpMetrics{}, errors.Wrap(err, "failed to create http_requests_total metric")
	}

	if metrics.requestDuration, err = meter.Float64Histogram(
		getMetricsPrefix(prefix, httpRequestDurationSeconds),
		metric.WithDescription("The HTTP request latencies in seconds"),
		metric.WithUnit("seconds"),
	); err != nil {
		return httpMetrics{}, errors.Wrap(err, "failed to create http_requests_duration_seconds metric")
	}

	return
}
