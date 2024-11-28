package rabbit

import (
	"context"

	"github.com/wagslane/go-rabbitmq"
	"github.com/xBlaz3kx/DevX/observability"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	Publisher *rabbitmq.Publisher
	obs       observability.Observability
	metrics   rabbitMetrics
}

func newPublisher(publisher *rabbitmq.Publisher, metrics rabbitMetrics, obs observability.Observability) Publisher {
	return Publisher{
		Publisher: publisher,
		obs:       obs.WithSpanKind(trace.SpanKindProducer),
		metrics:   metrics,
	}
}

func (pb *Publisher) Publish(ctx context.Context, topic string, message proto.Message, correlationID string, replyTopic Topic, optionFuncs ...PublishOpt) error {
	logger := pb.obs.Log().Ctx(ctx).With(zap.String("topic", topic), zap.String("correlationId", correlationID))

	// Apply options
	publisherOptions := newPublisherOptions()
	for _, optionFunc := range optionFuncs {
		optionFunc(publisherOptions)
	}

	// Get the headers
	headers := getPublisherHeaders(ctx, publisherOptions)

	// Marshall the payload
	payload, err := proto.Marshal(message)
	if err != nil {
		logger.Error("Error marshalling message", zap.Error(err))
		return err
	}

	// Publish the message
	err = pb.Publisher.Publish(
		payload,
		[]string{topic},
		rabbitmq.WithPublishOptionsExchange(string(CentralExchange)),
		rabbitmq.WithPublishOptionsContentType("application/vnd.google.protobuf"),
		rabbitmq.WithPublishOptionsCorrelationID(correlationID),
		rabbitmq.WithPublishOptionsHeaders(headers),
		rabbitmq.WithPublishOptionsReplyTo(string(replyTopic)),
	)
	if err != nil {
		logger.Error("Error publishing a message", zap.Error(err))
		return err
	}

	// Increment the number of messages published
	pb.metrics.IncrementMessagesPublished(topic)

	logger.With(zap.Any("headers", headers)).Debug("Published message")

	return nil
}

func getPublisherHeaders(ctx context.Context, publisherOptions *PublisherOptions) rabbitmq.Table {
	headers := make(rabbitmq.Table)
	for _, hv := range publisherOptions.headers {
		headers[string(hv.Key)] = hv.Value
	}

	// Add tracing headers if tracing is enabled
	if publisherOptions.tracing {
		traceHeaders := extractTraceFromContex(ctx)
		for key, value := range traceHeaders {
			headers[key] = value
		}
	}

	return headers
}
