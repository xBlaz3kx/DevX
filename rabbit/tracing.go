package rabbit

import (
	"context"

	"github.com/wagslane/go-rabbitmq"
	"go.opentelemetry.io/otel"
)

type TraceCarrier rabbitmq.Table

func (c TraceCarrier) Get(key string) string {
	v, ok := c[key]
	if !ok {
		return ""
	}
	return v.(string)
}

func (c TraceCarrier) Set(key string, value string) {
	c[key] = value
}

func (c TraceCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

func extractTraceFromContex(ctx context.Context) rabbitmq.Table {
	carrier := make(TraceCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return rabbitmq.Table(carrier)
}

// injectTraceFromHeaders
func injectTraceFromHeaders(ctx context.Context, carrier rabbitmq.Table) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, TraceCarrier(carrier))
}
