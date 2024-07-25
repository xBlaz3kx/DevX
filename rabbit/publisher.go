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
}

func newPublisher(publisher *rabbitmq.Publisher, obs observability.Observability) Publisher {
	return Publisher{
		Publisher: publisher,
		obs:       obs.WithSpanKind(trace.SpanKindProducer),
	}
}

func (pb *Publisher) Publish(ctx context.Context, topic string, message proto.Message, correlationID string, replyTopic Topic, optionFuncs ...func(*PublisherOptions)) error {
	logger := pb.obs.Log().Ctx(ctx)

	publisherOptions := newPublisherOptions()

	// Apply options
	for _, optionFunc := range optionFuncs {
		optionFunc(publisherOptions)
	}

	// Increment the number of messages published
	pb.obs.Metrics().IncrementMessagesPublished(ctx, topic)

	headers := InjectRabbitHeaders(ctx)

	for _, hv := range publisherOptions.headers {
		headers[string(hv.Key)] = hv.Value
	}

	// Marshall the payload
	payload, err := proto.Marshal(message)
	if err != nil {
		logger.With(
			zap.String("topic", topic),
			zap.String("correlationId", correlationID),
		).Error("Error marshalling message", zap.Error(err))
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

	logger.Debug("Published message",
		zap.String("topic", topic),
		zap.String("correlationId", correlationID),
		zap.Any("headers", headers),
	)

	return nil
}
