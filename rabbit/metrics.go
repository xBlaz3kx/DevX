package rabbit

import (
	"context"
	"fmt"

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
	rabbitConsumersTotal            = "rabbit_consumers"
	rabbitPublishersTotal           = "rabbit_publishers"

	attrQueueName = "queue_name"
)

type rabbitMetrics struct {
	messagesPublished    metric.Int64Counter
	messagesDelivered    metric.Int64Counter
	messagesAcknowledged metric.Int64Counter
	messagesRequeued     metric.Int64Counter
	messagesRejected     metric.Int64Counter
	consumers            metric.Int64Gauge
	publishers           metric.Int64Gauge
}

// Returns the metric name with the prefix
func getMetricsPrefix(prefix string, metric string) string {
	if prefix == "" {
		return metric
	}

	return fmt.Sprintf("%s_%s", prefix, metric)
}

// Initializes rabbit meters
func newRabbitMetrics(prefix string) (metrics rabbitMetrics, err error) {
	meter := otel.Meter("rabbit")

	if metrics.messagesPublished, err = meter.Int64Counter(
		getMetricsPrefix(prefix, rabbitMessagesPublishedTotal),
		metric.WithDescription("Total number of RabbitMQ messages published"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_published_total metric")
	}

	if metrics.messagesDelivered, err = meter.Int64Counter(
		getMetricsPrefix(prefix, rabbitMessagesDeliveredTotal),
		metric.WithDescription("Total number of RabbitMQ messages delivered"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_delivered_total metric")
	}

	if metrics.messagesAcknowledged, err = meter.Int64Counter(
		getMetricsPrefix(prefix, rabbitMessagesAcknowledgedTotal),
		metric.WithDescription("Total number of RabbitMQ acknowledged messages"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_acknowledged_total metric")
	}

	if metrics.messagesRequeued, err = meter.Int64Counter(
		getMetricsPrefix(prefix, rabbitMessagesRequeuedTotal),
		metric.WithDescription("Total number of RabbitMQ requeued messages"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_requeued_total metric")
	}

	if metrics.messagesRejected, err = meter.Int64Counter(
		getMetricsPrefix(prefix, rabbitMessagesRejectedTotal),
		metric.WithDescription("Total number of RabbitMQ rejected messages"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_messages_rejected_total metric")
	}

	if metrics.consumers, err = meter.Int64Gauge(
		getMetricsPrefix(prefix, rabbitConsumersTotal),
		metric.WithDescription("Total number of RabbitMQ consumers"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_consumers_total metric")
	}

	if metrics.publishers, err = meter.Int64Gauge(
		getMetricsPrefix(prefix, rabbitPublishersTotal),
		metric.WithDescription("Total number of publishers"),
	); err != nil {
		return rabbitMetrics{}, errors.Wrap(err, "failed to create rabbit_publishers_total metric")
	}

	return
}

func (m *rabbitMetrics) IncrementMessagesPublished(queueName string) {
	m.messagesPublished.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String(attrQueueName, queueName)),
	)
}

func (m *rabbitMetrics) IncrementMessagesDelivered(queueName string) {
	m.messagesDelivered.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String(attrQueueName, queueName)),
	)
}

func (m *rabbitMetrics) IncrementMessagesAcknowledged(queueName string) {
	m.messagesAcknowledged.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String(attrQueueName, queueName)),
	)
}

func (m *rabbitMetrics) IncrementMessagesRequeued(queueName string) {
	m.messagesRequeued.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String(attrQueueName, queueName)),
	)
}

func (m *rabbitMetrics) IncrementMessagesDiscarded(queueName string) {
	m.messagesRejected.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String(attrQueueName, queueName)),
	)
}

func (m *rabbitMetrics) IncrementConsumers(queueName string) {
	m.consumers.Record(context.Background(), 1,
		metric.WithAttributes(attribute.String(attrQueueName, queueName)),
	)
}

func (m *rabbitMetrics) DecrementConsumers(queueName string) {
	m.consumers.Record(context.Background(), 1,
		metric.WithAttributes(attribute.String(attrQueueName, queueName)),
	)
}
