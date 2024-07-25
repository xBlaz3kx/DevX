package rabbit

import (
	"context"

	"github.com/wagslane/go-rabbitmq"
	"go.opentelemetry.io/otel"
)

type HeadersCarrier rabbitmq.Table

func (c HeadersCarrier) Get(key string) string {
	v, ok := c[key]
	if !ok {
		return ""
	}
	return v.(string)
}

func (c HeadersCarrier) Set(key string, value string) {
	c[key] = value
}

func (c HeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

func InjectRabbitHeaders(ctx context.Context) rabbitmq.Table {
	carrier := make(HeadersCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return rabbitmq.Table(carrier)
}

func ExtractRabbitHeaders(ctx context.Context, carrier rabbitmq.Table) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, HeadersCarrier(carrier))
}
