package rabbit

import (
	"context"

	"github.com/wagslane/go-rabbitmq"
)

type client struct {
	responseChannel chan ReplyResponse

	// expectedResponseNumber represents the number of responses
	// expected for this client
	expectedResponseNumber int
}

type ReplyPool struct {
	Request  chan ReplyRequest
	Response chan ReplyResponse
	Cancel   chan string
	Clients  map[string]*client
}

type ReplyRequest struct {
	CorrelationId string
	RequestChan   chan ReplyResponse
	// ExpectedResponsesNr represents the number of responses
	// expected for this client
	ExpectedResponsesNr int
}

type ReplyResponse struct {
	CorrelationId string
	Body          []byte
	Error         bool
	Headers       map[string]any
}

// NewReplyPool creates and returns a new ReplyPool
func NewReplyPool(bufferSize int) ReplyPool {
	return ReplyPool{
		Request:  make(chan ReplyRequest, bufferSize),
		Response: make(chan ReplyResponse, bufferSize),
		Clients:  make(map[string]*client),
	}
}

// Start starts the ReplyPool routine
func (rp *ReplyPool) start() {
	for {
		select {
		case req := <-rp.Request:
			rp.Clients[req.CorrelationId] = &client{
				responseChannel:        req.RequestChan,
				expectedResponseNumber: req.ExpectedResponsesNr,
			}
		case res := <-rp.Response:
			client, ok := rp.Clients[res.CorrelationId]
			if ok {
				responseChannel := client.responseChannel
				client.expectedResponseNumber--

				if client.expectedResponseNumber <= 0 {
					delete(rp.Clients, res.CorrelationId)
				}

				responseChannel <- res
			}
		case cnc := <-rp.Cancel:
			delete(rp.Clients, cnc)
		}
	}
}

// NewReplyConsumer creates a new reply queue consumer
func newReplyConsumer(consumer Consumer, responseChannel chan<- ReplyResponse, topic Topic, routines int) (*TopicConsumer, error) {
	return consumer.newConsumer(consumer.exchange, topic, string(topic),
		func(ctx context.Context, d rabbitmq.Delivery) (action rabbitmq.Action) {
			response := ReplyResponse{
				CorrelationId: d.CorrelationId,
				Body:          d.Body,
				Error:         d.Headers["error"].(bool),
				Headers:       d.Headers,
			}
			responseChannel <- response
			return rabbitmq.Ack
		},
		false,    // Reply queues are not durable as they are made per-instance
		routines, // Run multiple routines for consumer
	)
}
