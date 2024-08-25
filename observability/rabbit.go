package observability

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	rabbitMessagesPublishedTotal    = "rabbit_messages_published_total"
	rabbitMessagesDeliveredTotal    = "rabbit_messages_delivered_total"
	rabbitMessagesAcknowledgedTotal = "rabbit_messages_acknowledged_total"
	rabbitMessagesRequeuedTotal     = "rabbit_messages_requeued_total"
	rabbitMessagesRejectedTotal     = "rabbit_messages_rejected_total"
	rabbitConsumersTotal            = "rabbit_consumers_total"

	attrQueueName = "queue_name"
)

type rabbitMetrics struct {
	messagesPublished    metric.Int64Counter
	messagesDelivered    metric.Int64Counter
	messagesAcknowledged metric.Int64Counter
	messagesRequeued     metric.Int64Counter
	messagesRejected     metric.Int64Counter
	consumers            metric.Int64ObservableGauge
}

var (
	consumerCount int64 = 0
)

// Initializes rabbit meters
func newRabbitMetrics() (metrics rabbitMetrics, err error) {
	meter := otel.Meter("rabbit")

	if metrics.messagesPublished, err = meter.Int64Counter(
		rabbitMessagesPublishedTotal,
		metric.WithDescription("Total number of RabbitMQ messages published"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_published_total metric")
	}

	if metrics.messagesDelivered, err = meter.Int64Counter(
		rabbitMessagesDeliveredTotal,
		metric.WithDescription("Total number of RabbitMQ messages delivered"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_delivered_total metric")
	}

	if metrics.messagesAcknowledged, err = meter.Int64Counter(
		rabbitMessagesAcknowledgedTotal,
		metric.WithDescription("Total number of RabbitMQ acknowledged messages"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_acknowledged_total metric")
	}

	if metrics.messagesRequeued, err = meter.Int64Counter(
		rabbitMessagesRequeuedTotal,
		metric.WithDescription("Total number of RabbitMQ requeued messages"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_requeued_total metric")
	}

	if metrics.messagesRejected, err = meter.Int64Counter(
		rabbitMessagesRejectedTotal,
		metric.WithDescription("Total number of RabbitMQ rejected messages"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_rejected_total metric")
	}

	if metrics.consumers, err = meter.Int64ObservableGauge(
		rabbitConsumersTotal,
		metric.WithDescription("Total number of RabbitMQ consumers"),
		metric.WithInt64Callback(func(ctx context.Context, io metric.Int64Observer) error {
			io.Observe(consumerCount)
			return nil
		}),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_consumers_total metric")
	}

	return
}

func (m *Metrics) IncrementMessagesPublished(ctx context.Context, queueName string) {
	if m == nil {
		return
	}

	m.rabbit.messagesPublished.Add(
		context.Background(), 1,
		metric.WithAttributes(
			attribute.String(attrQueueName, queueName),
		),
	)
}

func (m *Metrics) IncrementMessagesDelivered(queueName string) {
	if m == nil {
		return
	}

	m.rabbit.messagesDelivered.Add(
		context.Background(), 1,
		metric.WithAttributes(
			attribute.String(attrQueueName, queueName),
		),
	)
}

func (m *Metrics) IncrementMessagesAcknowledged(queueName string) {
	if m == nil {
		return
	}

	m.rabbit.messagesAcknowledged.Add(
		context.Background(), 1,
		metric.WithAttributes(
			attribute.String(attrQueueName, queueName),
		),
	)
}

func (m *Metrics) IncrementMessagesRequeued(queueName string) {
	if m == nil {
		return
	}

	m.rabbit.messagesRequeued.Add(
		context.Background(), 1,
		metric.WithAttributes(
			attribute.String(attrQueueName, queueName),
		),
	)
}

func (m *Metrics) IncrementMessagesDiscarded(queueName string) {
	if m == nil {
		return
	}

	m.rabbit.messagesRejected.Add(
		context.Background(), 1,
		metric.WithAttributes(
			attribute.String(attrQueueName, queueName),
		),
	)
}

// todo underlying implementation should use metrics package instead?
func (m *Metrics) IncrementConsumers(queueName string) {
	if m == nil {
		return
	}
	atomic.AddInt64(&consumerCount, 1)
}

func (m *Metrics) DecrementConsumers(queueName string) {
	if m == nil || consumerCount == 0 {
		return
	}
	atomic.AddInt64(&consumerCount, -1)
}
