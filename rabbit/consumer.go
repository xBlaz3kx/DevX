package rabbit

import (
	"context"

	"github.com/wagslane/go-rabbitmq"
	"github.com/xBlaz3kx/DevX/observability"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type HandlerFunc func(ctx context.Context, d rabbitmq.Delivery) (action rabbitmq.Action)

type ConsumerFactory struct {
	connection *rabbitmq.Conn
	opts       ConsumerOpts
	exchange   Exchange

	// Observability
	obs     observability.Observability
	metrics rabbitMetrics
}

// NewConsumerFactory creates a new consumer factory for the given exchange
// Opts will be applied to every consumer created by this factory and can be overridden by the consumer itself
func NewConsumerFactory(connection *rabbitmq.Conn, exchange Exchange, metrics rabbitMetrics, obs observability.Observability, opts ...ConsumerOpt) ConsumerFactory {
	// Apply opts
	consumerOptions := newConsumerOptions()
	for _, opt := range opts {
		opt(&consumerOptions)
	}

	consumer := ConsumerFactory{
		connection: connection,
		exchange:   exchange,
		obs:        obs.WithSpanKind(trace.SpanKindConsumer),
		opts:       consumerOptions,
		metrics:    metrics,
	}

	return consumer
}

// NewConsumer creates a new consumer for given exchange, topic and handler function, the returned consumer should only be used for disconnecting
func (cm *ConsumerFactory) NewConsumer(exchange Exchange, topic Topic, queueName string, handler HandlerFunc, durable bool, opts ...ConsumerOpt) (*rabbitmq.Consumer, error) {
	logger := cm.obs.Log().With(
		zap.String("exchange", string(exchange)),
		zap.String("topic", string(topic)),
		zap.String("queueName", queueName),
		zap.Bool("durable/quorum", durable),
	)

	// Override default options with the given options
	consumerOptions := cm.opts
	for _, opt := range opts {
		opt(&consumerOptions)
	}

	// Set up the consumer options
	options := []func(*rabbitmq.ConsumerOptions){
		rabbitmq.WithConsumerOptionsExchangeName(string(exchange)),
		rabbitmq.WithConsumerOptionsConcurrency(consumerOptions.routines),
		rabbitmq.WithConsumerOptionsRoutingKey(string(topic)),
	}

	if durable {
		options = append(options, rabbitmq.WithConsumerOptionsQueueDurable, rabbitmq.WithConsumerOptionsQueueQuorum)
	}

	// Set up the handler for the message
	rabbitHandler := func(d rabbitmq.Delivery) rabbitmq.Action {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), cm.opts.eventTimeout)
		defer cancel()

		ctx := injectTraceFromHeaders(timeoutCtx, rabbitmq.Table(d.Headers))
		logger.Debug("Received message on the consumer", zap.Any("headers", d.Headers))

		// Increment the number of messages delivered for the given topic
		cm.metrics.IncrementMessagesDelivered(string(topic))

		// Call the handler function
		action := handler(ctx, d)

		// Depending on the response, increment the appropriate metric
		switch action {
		case rabbitmq.Ack:
			cm.metrics.IncrementMessagesAcknowledged(string(topic))
		case rabbitmq.NackRequeue:
			cm.metrics.IncrementMessagesRequeued(string(topic))
		case rabbitmq.NackDiscard:
			cm.metrics.IncrementMessagesDiscarded(string(topic))
		default:
			// Do nothing
		}

		return action
	}

	consumer, err := rabbitmq.NewConsumer(
		cm.connection,
		queueName,
		options...,
	)
	if err != nil {
		return nil, err
	}

	go func() {
		err = consumer.Run(rabbitHandler)
		if err != nil {
			logger.Panic("Error running rabbit consumer", zap.Error(err))
		}
	}()

	return consumer, nil
}
