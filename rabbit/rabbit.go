package rabbit

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/wagslane/go-rabbitmq"
	"github.com/xBlaz3kx/DevX/observability"
	grpc "github.com/xBlaz3kx/DevX/proto"
	"go.uber.org/zap"
)

var ErrResponse = errors.New("responded with an error")

// Configuration AMQP basic configuration for the message bus
type Configuration struct {
	// Address is the address for connecting to the RabbitMQ instance
	Address string `validate:"required" json:"address" yaml:"address"`

	// Username for authentication to the RabbitMQ instance
	Username string `json:"username" yaml:"username"`

	// Password for authentication to the RabbitMQ instance
	Password string `json:"password" yaml:"password"`
}

type Rabbit struct {
	Consumer         Consumer
	Publisher        PublisherPool
	replyPool        ReplyPool
	connections      []*rabbitmq.Conn
	connectionString string
	configuration    Configuration
	options          *Options
	Exchange         Exchange
	replyTopic       Topic
	obs              observability.Observability
}

func NewError(errorMessage string, errorCode grpc.ErrorCode) *grpc.Error {
	return &grpc.Error{
		Code:    errorCode,
		Message: errorMessage,
	}
}

// New creates and returns a new rabbit client with given configuration
func New(configuration Configuration, serviceExchange Exchange, obs observability.Observability, optionFuncts ...func(*Options)) (*Rabbit, error) {
	// Create reply topic
	instanceHostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	replyTopic := NewTopic(serviceExchange).AddWord(replyBase).AddWord(TopicWord(instanceHostname)).Build()

	// Setup defaults
	options := newRabbitOptions()

	// Setup the logger
	optionFuncts = append(optionFuncts)

	// Apply options
	for _, optionFunc := range optionFuncts {
		optionFunc(options)
	}

	logger := obs.Log().With(
		zap.String("exchange", string(serviceExchange)),
		zap.String("replyTopic", string(replyTopic)),
		zap.Int("replyConsumers", options.replyConsumers),
		zap.Int("publishers", options.publishers),
	)
	logger.Debug("Starting Rabbitmq")

	client := &Rabbit{
		options:          options,
		Exchange:         serviceExchange,
		replyTopic:       replyTopic,
		configuration:    configuration,
		obs:              obs,
		connectionString: fmt.Sprintf("amqp://%v:%v@%v", configuration.Username, configuration.Password, configuration.Address),
	}

	// Create a TCP connection for the consumers
	conn, err := client.createConnection()
	if err != nil {
		return nil, err
	}

	// Create a Consumer
	client.Consumer = newConsumer(conn, serviceExchange, obs)

	// Create a reply pool and start it in a dedicated routine
	client.replyPool = NewReplyPool(30)
	go client.replyPool.start()

	// Start a reply consumer
	_, err = newReplyConsumer(client.Consumer, client.replyPool.Response, replyTopic, options.replyConsumers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create reply consumer")
	}

	// Start a publisher pool
	poolPublishers, err := client.createPublishers(options.publishers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create publishers")
	}

	client.Publisher = newPublisherPool(poolPublishers, client.replyPool, serviceExchange, replyTopic, obs)
	go client.Publisher.start()

	logger.Info("Rabbit service started")

	return client, nil
}

// Connect connects the rabbit client to rabbitmq server
func (c *Rabbit) createConnection() (*rabbitmq.Conn, error) {
	c.obs.Log().With(zap.String("address", c.configuration.Address)).Debug("Creating a rabbit connection")

	conn, err := rabbitmq.NewConn(c.connectionString, rabbitmq.WithConnectionOptionsReconnectInterval(time.Second))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a rabbit connection")
	}

	c.connections = append(c.connections, conn)
	return conn, nil
}

// Disconnect disconnects all rabbit connections
func (c *Rabbit) Disconnect() error {
	c.obs.Log().Debug("Disconnecting all rabbit connections")

	for _, c := range c.connections {
		err := c.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// createPublishers creates the desired number of publishers
func (c *Rabbit) createPublishers(number int) ([]Publisher, error) {
	publishers := []Publisher{}

	for i := 0; i < number; i++ {
		conn, err := c.createConnection()
		if err != nil {
			return nil, err
		}

		publisher, err := rabbitmq.NewPublisher(conn, rabbitmq.WithPublisherOptionsLogger(c.options.logger))
		if err != nil {
			return nil, err
		}

		publishers = append(publishers, newPublisher(publisher, c.obs))
	}

	return publishers, nil
}

func (c *Rabbit) Pass() bool {
	// todo need to modify the underlying library to expose IsClosed attribute
	return true
}

func (c *Rabbit) Name() string {
	return "rabbitmq"
}
