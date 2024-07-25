package rabbit

import (
	"github.com/wagslane/go-rabbitmq"
)

type Options struct {
	logger         rabbitmq.Logger
	publishers     int
	replyConsumers int
}

func newRabbitOptions() *Options {
	return &Options{
		publishers:     1,
		replyConsumers: 1,
	}
}

func WithLogger(logger rabbitmq.Logger) func(options *Options) {
	return func(options *Options) {
		options.logger = logger
	}
}

// WithMultiplePublishers sets the number of routines running a publisher, each has it's own TCP connection
func WithMultiplePublishers(number int) func(options *Options) {
	return func(options *Options) {
		options.publishers = number
	}
}

// WithConcurrentReplyConsumer sets the number of go routines that will handle the reply queue
func WithConcurrentReplyConsumer(number int) func(options *Options) {
	return func(options *Options) {
		options.replyConsumers = number
	}
}

type PublisherOptions struct {
	headers []HeaderValue
}

func newPublisherOptions() *PublisherOptions {
	return &PublisherOptions{}
}

func WithHeader(header []HeaderValue) func(options *PublisherOptions) {
	return func(options *PublisherOptions) {
		options.headers = append(options.headers, header...)
	}
}
