package observability

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type httpMetrics struct {
	requestsTotal   metric.Int64Counter
	requestDuration metric.Float64Histogram
	errorsTotal     metric.Int64Counter
}

const (
	httpRequestsTotal          = "http_requests_total"
	httpRequestDurationSeconds = "http_requests_duration_seconds"
	httpErrorsTotal            = "http_errors_total"

	attrStatus = "status"
	attrPath   = "path"
	attrMethod = "method"
	attrUserId = "user_id"
)

// Initializes http meters
func newHttpMetrics() (metrics httpMetrics, err error) {
	meter := otel.Meter("http")

	if metrics.requestsTotal, err = meter.Int64Counter(
		httpRequestsTotal,
		metric.WithDescription("Total number of HTTP requests"),
	); err != nil {
		return httpMetrics{}, errors.Wrap(err, "failed to create http_requests_total metric")
	}

	if metrics.requestDuration, err = meter.Float64Histogram(
		httpRequestDurationSeconds,
		metric.WithDescription("The HTTP request latencies in seconds"),
		metric.WithUnit("seconds"),
	); err != nil {
		return httpMetrics{}, errors.Wrap(err, "failed to create http_requests_duration_seconds metric")
	}

	if metrics.errorsTotal, err = meter.Int64Counter(
		httpErrorsTotal,
		metric.WithDescription("Total number of HTTP errors"),
	); err != nil {
		return httpMetrics{}, errors.Wrap(err, "failed to create http_errors_total metric")
	}

	return
}

// Middleware returns a gin middleware that records metrics for each HTTP request.
func (m *Metrics) Middleware(tracingEnabled bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		if m == nil {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		startTime := time.Now()

		c.Next()

		attrs := []attribute.KeyValue{
			attribute.Int(attrStatus, c.Writer.Status()),
			attribute.String(attrPath, c.FullPath()),
			attribute.String(attrMethod, c.Request.Method),
		}

		if tracingEnabled {
			span := trace.SpanFromContext(ctx)
			// Add trace ID and span ID to the metrics if the span is being recorded,
			// forming an exemplar
			if span.IsRecording() {
				attrs = append(
					attrs,
					attribute.String("trace_id", span.SpanContext().TraceID().String()),
					attribute.String("span_id", span.SpanContext().SpanID().String()),
				)
			}
		}

		if c.GetHeader("X-User") != "" {
			attrs = append(attrs, attribute.String(attrUserId, c.GetHeader("X-User")))
		}

		attributes := metric.WithAttributes(attrs...)

		m.http.requestDuration.Record(ctx, time.Since(startTime).Seconds(), attributes)
		m.http.requestsTotal.Add(ctx, 1, attributes)

		if c.Writer.Status() >= 400 {
			m.http.errorsTotal.Add(ctx, 1, attributes)
		}
	}
}
