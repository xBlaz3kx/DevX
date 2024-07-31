package rabbit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type exampleLogger struct {
}

func (e *exampleLogger) Fatalf(s string, i ...interface{}) {
}

func (e *exampleLogger) Errorf(s string, i ...interface{}) {
}

func (e *exampleLogger) Warnf(s string, i ...interface{}) {
}

func (e *exampleLogger) Infof(s string, i ...interface{}) {
}

func (e *exampleLogger) Debugf(s string, i ...interface{}) {
}

func TestOptionsWithLogger(t *testing.T) {
	options := newRabbitOptions()

	WithLogger(nil)(options)

	assert.Nil(t, options.logger)

	logger := &exampleLogger{}
	WithLogger(logger)(options)

	assert.NotNil(t, options.logger)
	assert.Equal(t, logger, options.logger)
}

func TestOptionsWithMultiplePublishers(t *testing.T) {
	options := newRabbitOptions()

	WithMultiplePublishers(1)(options)

	assert.Equal(t, 1, options.publishers)

	WithMultiplePublishers(2)(options)

	assert.Equal(t, 2, options.publishers)
}

func TestOptionsWithConcurrentReplyConsumer(t *testing.T) {
	options := newRabbitOptions()

	WithConcurrentReplyConsumer(1)(options)

	assert.Equal(t, 1, options.replyConsumers)

	WithConcurrentReplyConsumer(2)(options)

	assert.Equal(t, 2, options.replyConsumers)
}

func TestPublisherOptionsWithHeader(t *testing.T) {
	publisherOpts := newPublisherOptions()

	WithHeader([]HeaderValue{})(publisherOpts)

	assert.Equal(t, 0, len(publisherOpts.headers))

	WithHeader([]HeaderValue{{Key: "key", Value: "value"}})(publisherOpts)
	assert.Equal(t, 1, len(publisherOpts.headers))
	assert.EqualValues(t, "key", publisherOpts.headers[0].Key)
	assert.EqualValues(t, "value", publisherOpts.headers[0].Value)
	//assert.IsType(t, HeaderKey{}, publisherOpts.headers[0].Key)

	WithHeader([]HeaderValue{{Key: "key2", Value: "value2"}})(publisherOpts)
	assert.Equal(t, 2, len(publisherOpts.headers))
	assert.EqualValues(t, "key2", publisherOpts.headers[1].Key)
	assert.EqualValues(t, "value2", publisherOpts.headers[1].Value)
}
