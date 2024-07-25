package rabbit

import (
	"context"

	"github.com/wagslane/go-rabbitmq"
	"github.com/xBlaz3kx/DevX/observability"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	Consumer struct {
		connection *rabbitmq.Conn
		exchange   Exchange
		obs        observability.Observability
	}

	TopicConsumer struct {
		topic    Topic
		consumer *rabbitmq.Consumer
	}

	HandlerFunc func(ctx context.Context, d rabbitmq.Delivery) (action rabbitmq.Action)
)

// NewConsumer creates a new consumer for the given exchange
func newConsumer(connection *rabbitmq.Conn, exchange Exchange, obs observability.Observability) Consumer {
	return Consumer{
		connection: connection,
		exchange:   exchange,
		obs:        obs.WithSpanKind(trace.SpanKindConsumer),
	}
}

// NewServiceConsumer creates a new service topic consumer for given topic and handler function
// If durable is set to true, the queue will be of quorum type which is durable and persists through restarts
// routines sets the number of go routines handling the consumer, 1 should suffice in general
// returns the consumer and error, only initialize them if necessary
// the returned consumer should only be used for disconnecting
func (cm *Consumer) NewServiceConsumer(topic Topic, handler HandlerFunc, durable bool, routines int) (*TopicConsumer, error) {
	queueName := string(topic)
	return cm.newConsumer(cm.exchange, topic, queueName, handler, durable, routines)
}

func (cm *Consumer) NewChargePointConsumer(topic Topic, handler HandlerFunc) (*TopicConsumer, error) {
	queueName := string(topic)
	return cm.newConsumer(cm.exchange, topic, queueName, handler, false, 1)
}

// NewServiceConsumer creates a new service topic consumer for given topic and handler function
// If durable is set to true, the queue will be of quorum type which is durable and persists through restarts
// routines sets the number of go routines handling the consumer, 1 should suffice in general
// returns the consumer and error, only initialize them if necessary
// the returned consumer should only be used for disconnecting
func (cm *Consumer) NewService1Consumer(topic Topic, handler HandlerFunc) (*TopicConsumer, error) {
	queueName := string(topic)
	return cm.newConsumer(cm.exchange, topic, queueName, handler, true, 1)
}

// newConsumer creates a new consumer for given exchange, topic and handler function, the returned consumer should only be used for disconnecting
func (cm *Consumer) newConsumer(exchange Exchange, topic Topic, queueName string, handler HandlerFunc, durable bool, routines int) (*TopicConsumer, error) {
	logger := cm.obs.Log().With(
		zap.String("exchange", string(exchange)),
		zap.String("topic", string(topic)),
		zap.String("queueName", queueName),
		zap.Int("routines", routines),
		zap.Bool("durable/quorum", durable),
	)

	// Set up the consumer options
	options := []func(*rabbitmq.ConsumerOptions){
		rabbitmq.WithConsumerOptionsExchangeName(string(exchange)),
		rabbitmq.WithConsumerOptionsConcurrency(routines),
		rabbitmq.WithConsumerOptionsRoutingKey(string(topic)),
	}

	if durable {
		options = append(options, rabbitmq.WithConsumerOptionsQueueDurable, rabbitmq.WithConsumerOptionsQueueQuorum)
	}

	// Set up the handler for the message
	rabbitHandler := func(d rabbitmq.Delivery) rabbitmq.Action {
		ctx := ExtractRabbitHeaders(context.Background(), rabbitmq.Table(d.Headers))

		logger.Debug("Received message on the consumer",
			zap.String("exchange", string(cm.exchange)),
			zap.Int("routines", routines),
			zap.Bool("durable/quorum", durable),
			zap.String("topic", string(topic)),
			zap.String("queueName", queueName),
			zap.Any("headers", d.Headers),
		)

		// Increment the number of messages delivered for the given topic
		cm.obs.Metrics().IncrementMessagesDelivered(string(topic))

		// Call the handler function
		action := handler(ctx, d)

		// Depending on the response, increment the appropriate metric
		switch action {
		case rabbitmq.Ack:
			cm.obs.Metrics().IncrementMessagesAcknowledged(string(topic))
		case rabbitmq.NackRequeue:
			cm.obs.Metrics().IncrementMessagesRequeued(string(topic))
		case rabbitmq.NackDiscard:
			cm.obs.Metrics().IncrementMessagesDiscarded(string(topic))
		}

		return action
	}

	consumer, err := rabbitmq.NewConsumer(
		cm.connection,
		queueName,
		options...,
	)
	if err != nil {
		logger.Panic("Error creating rabbit consumer", zap.Error(err))
		return nil, err
	}

	go func() {
		err = consumer.Run(rabbitHandler)
		if err != nil {
			logger.Panic("Error running rabbit consumer", zap.Error(err))
		}
	}()

	logger.Debug("Started rabbit consumer")
	return &TopicConsumer{
		topic:    topic,
		consumer: consumer,
	}, nil
}

// NewNotificationConsumer creates a new consumer of notifications for given topic, with routing key and handler function, the returned consumer should only be used for disconnecting
// - routines sets the number of go routines handling the consumer, 1 should suffice in general
// - returns the consumer and error, only initialize them if necessary
// - to achieve load balancing, queueName should be set in the form of <consumer-service>.<topic>
func (cm *Consumer) NewNotificationConsumer(topic Topic, queueName string, handler HandlerFunc, routines int) (*TopicConsumer, error) {
	return cm.newConsumer(GlobalNotificationExchange, topic, queueName, handler, true, routines)
}

// Close closes the consumer
func (tc *TopicConsumer) Close() {
	tc.consumer.Close()
}
