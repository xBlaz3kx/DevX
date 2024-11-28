package rabbit

import (
	"context"

	"github.com/google/uuid"
	"github.com/xBlaz3kx/DevX/observability"
	grpc "github.com/xBlaz3kx/DevX/proto"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type PublisherPool struct {
	publishers []Publisher
	request    chan *PublishRequest
	roundRobin int
	replyPool  ReplyPool
	exchange   Exchange
	replyTopic Topic
	obs        observability.Observability
}

type PublishRequest struct {
	Ctx             context.Context
	Topic           Topic
	CorrelationId   string
	Message         proto.Message
	Options         []PublishOpt
	ResponseChannel chan error
}

// NewPublisherPool creates a new publisher pool that handles all publishing for the service
// It is routine-safe
func newPublisherPool(publishers []Publisher, replyPool ReplyPool, exchange Exchange, replyTopic Topic, obs observability.Observability) PublisherPool {
	publisherPool := PublisherPool{
		publishers: publishers,
		request:    make(chan *PublishRequest, 30),
		roundRobin: 0,
		replyPool:  replyPool,
		exchange:   exchange,
		replyTopic: replyTopic,
		obs:        obs.WithSpanKind(trace.SpanKindProducer),
	}
	return publisherPool
}

// Start starts the PublisherPool routine
func (pp *PublisherPool) start() {
	for req := range pp.request {
		req.ResponseChannel <- pp.publishers[pp.roundRobin].Publish(req.Ctx, string(req.Topic), req.Message, req.CorrelationId, pp.replyTopic, req.Options...)
		if pp.roundRobin == len(pp.publishers)-1 {
			pp.roundRobin = 0
		} else {
			pp.roundRobin++
		}
	}
}

// Publish publishes a rabbit message
// Returns an error, only initialize it if needed, error already logged
func (pp *PublisherPool) Publish(ctx context.Context, topic Topic, message proto.Message, options ...PublishOpt) error {
	correlationId := uuid.New().String()
	// Create a channel for the response and close it when done
	errChan := make(chan error, 1)
	defer close(errChan)

	publishRequest := &PublishRequest{
		Ctx:             ctx,
		Topic:           topic,
		CorrelationId:   correlationId,
		Message:         message,
		Options:         options,
		ResponseChannel: errChan,
	}

	pp.request <- publishRequest

	err := waitError(ctx, publishRequest.ResponseChannel)
	if err != nil {
		pp.obs.Log().Error(
			"Unable to publish rabbit message",
			zap.Error(err),
			zap.String("correlationId", correlationId),
			zap.String("topic", string(topic)),
		)
	}
	return err
}

// Respond publishes a response to a rabbit message
// Set isError to true if the reply is an error, otherwise pass false to indicate valid response
// Returns an error, only initialize it if necessary
func (pp *PublisherPool) Respond(ctx context.Context, correlationID string, topic Topic, message proto.Message, options ...PublishOpt) error {
	return pp.respond(ctx, correlationID, topic, message, false, options...)
}

func (pp *PublisherPool) RespondWithError(ctx context.Context, correlationID string, topic Topic, message *grpc.Error, options ...PublishOpt) error {
	return pp.respond(ctx, correlationID, topic, message, true, options...)
}

// RespondWithHeader publishes a response to a rabbit message with additional header values.
func (pp *PublisherPool) respond(ctx context.Context, correlationID string, topic Topic, message proto.Message, isError bool, options ...PublishOpt) error {
	errChan := make(chan error, 1)
	defer close(errChan)

	header := NewHeader().WithError(isError).Build()
	options = append(options, WithPublisherHeader(header))

	publishRequest := &PublishRequest{
		Ctx:             ctx,
		Topic:           topic,
		CorrelationId:   correlationID,
		Message:         message,
		Options:         options,
		ResponseChannel: errChan,
	}

	if isError {
		errPayload := message.(*grpc.Error)
		pp.obs.Log().With(
			zap.String("error", errPayload.Message),
			zap.Int32("code", int32(errPayload.Code)),
		).Debug("Responding with error")
	}

	pp.request <- publishRequest

	err := waitError(ctx, publishRequest.ResponseChannel)
	if err != nil {
		pp.obs.Log().With(
			zap.String("correlationId", correlationID),
			zap.String("topic", string(topic)),
			zap.Error(err),
		).Debug("Rabbit unable to publish a response")
	}

	return err
}

// PublishRPC publishes a RPC message and waits for the reply
func (pp *PublisherPool) PublishRPC(ctx context.Context, topic Topic, message proto.Message, options ...PublishOpt) ([]byte, error) {
	replyChannel, err := pp.PublishRPCWithMultipleResponses(ctx, topic, message, 1, options...)
	if err != nil {
		return nil, err
	}

	return waitReply(ctx, replyChannel)
}

func (pp *PublisherPool) PublishRPCWithMultipleResponses(ctx context.Context, topic Topic, message proto.Message, nrResponses int, options ...PublishOpt) (chan ReplyResponse, error) {
	correlationId := uuid.New().String()
	// Create a channel for the response and close it when done
	errChan := make(chan error, 1)
	defer close(errChan)

	publishRequest := &PublishRequest{
		Ctx:             ctx,
		Topic:           topic,
		CorrelationId:   correlationId,
		Message:         message,
		Options:         options,
		ResponseChannel: errChan,
	}

	// Send a reply request to replyPool
	// replyChannel is buffered to prevent blocking the replyPool
	// if the response is not being read during sending.
	replyChannel := make(chan ReplyResponse, nrResponses)
	replyRequest := ReplyRequest{
		CorrelationId:       correlationId,
		RequestChan:         replyChannel,
		ExpectedResponsesNr: nrResponses,
	}

	pp.replyPool.Request <- replyRequest

	pp.request <- publishRequest

	err := waitError(ctx, publishRequest.ResponseChannel)
	if err != nil {
		pp.obs.Log().With(
			zap.String("topic", string(topic)),
			zap.Error(err),
		).Debug("Rabbit unable to publish a RPC request")
		// Publishing failed, cancel reply request
		pp.replyPool.Cancel <- correlationId
	}

	return replyChannel, err
}

// WaitError waits for a possible error and handles timeout
func waitError(ctx context.Context, errChan chan error) error {
	for {
		select {
		case d := <-errChan:
			return d
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// WaitReply waits for a reply for given time and returns an error if it times out
func waitReply(ctx context.Context, replyChannel chan ReplyResponse) ([]byte, error) {
	for {
		select {
		case d := <-replyChannel:
			if d.Error {
				return d.Body, ErrResponse
			}

			return d.Body, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
