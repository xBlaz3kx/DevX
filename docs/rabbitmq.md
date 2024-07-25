# Internal RabbitMQ library usage examples

## Initializing the library

In the main function of the service, initialize the rabbit library with the following code:
The `rabbitConfig` should be a struct containing the configuration for the rabbit client. The
`rabbitTypes.ChargePointExchange` is the exchange that the client will be using.

```go
    // Create and run the rabbit client
rabbitConfiguration, err := rabbit.GetConfiguration(rabbitConfig, rabbitTypes.Topic("Example"))
if err != nil {
obs.Log().Fatal("Could not create rabbit configuration", zap.Error(err))
}

rb, err := rabbit.NewRabbit(
rabbitConfiguration,
obs,
rabbit.WithConcurrentReplyConsumer(2),
rabbit.WithMultiplePublishers(2),
)
if err != nil {
obs.Log().Fatal("Could not create rabbit client", zap.Error(err))
}
```

## Creating a consumer

Here is an example on how to create a new consumer for a specific topic. In the consumer, only service calls should be
made. The consumer should not contain any business logic.

```go

// StatusNotificationConsumer consumes status notification messages
func (a *Api) StatusNotificationConsumer() {
handlerFunc := func (ctx context.Context, d rabbitmq.Delivery) (action rabbitmq.Action) {
ctx, end, logger := a.obs.LogSpan(ctx, "api.rabbit.StatusNotificationConsumer")
defer end()

ctx, cancel := context.WithTimeout(ctx, time.Second*10)
defer cancel()

request := &grpc.StatusNotification{}
if err := proto.Unmarshal(d.Body, request); err != nil {
return rabbitmq.Ack
}

logger = logger.With()

if err := a.chargePointService.StatusNotification(ctx, request); err != nil {
logger.Error("Unable to process status notification", zap.Error(err))
}

return rabbitmq.Ack
}

topic := rabbit.NewTopic(rabbitTypes.GlobalNotificationExchange).AddWord(rabbitTypes.StatusNotification).Build()
queueName := rabbit.NewTopic(a.rabbit.Exchange).AddWord(rabbitTypes.StatusNotification).Build().String()

a.rabbit.Consumer.NewNotificationConsumer(topic, queueName, handlerFunc, 1)
}

```

## Sending a notification

Here is an example on how to send a notification to the charge point management service or any service, subscribed to
the `GlobalNotificationExchange` exchange.

```go
// StatusNotification sends a status notification update
func StatusNotification(ctx context.Context, request *grpc.StatusNotification) error {
ctx, end := obs.Span(
ctx, "rabbit.StatusNotification"
)
defer end()

topic := rabbit.NewTopic(rabbitTypes.GlobalNotificationExchange).AddWord(rabbitTypes.StatusNotification).Build()
return rabbit.Publisher.Publish(ctx, topic, request)
}

```

## Responding to a request

In the consumer, you can respond to a request with the following code:

```go

// Respond with a response
if err := rabbit.Publisher.Respond(ctx, d.CorrelationId, rabbit.Topic(d.ReplyTo), response); err != nil {
logger.Error("Unable to respond to rabbit request", zap.Error(err))
}

```

## Error handling

Responding with an error to a request:

```go
response := rabbitTypes.NewError(err.Error(), grpc.ErrorCode_PayloadError)

// Respond with an error
if err := rabbit.Publisher.RespondWithError(ctx, d.CorrelationId, rabbit.Topic(d.ReplyTo), response); err != nil {
logger.Error("Unable to respond to rabbit request", zap.Error(err))
}
```