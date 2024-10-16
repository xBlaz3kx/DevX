package mqtt

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tavsec/gin-healthcheck/checks"
	"github.com/xBlaz3kx/DevX/configuration"
	"github.com/xBlaz3kx/DevX/observability"
)

const (
	MqttVersion5 = "v5"
	MqttVersion3 = "v3"
)

// Configuration is the configuration for the MQTT client
type Configuration struct {
	Version  string            `yaml:"version"`
	Address  string            `validate:"required,url" json:"address" yaml:"address"`
	Username string            `yaml:"username"`
	Password string            `yaml:"password"`
	ClientId string            `validate:"required" yaml:"clientId"`
	TLS      configuration.TLS `validate:"required" yaml:"tls"`
}

type Handler func(client Client, topicIds []string, payloadId uint16, payload interface{}, err error)

// Client is an interface wrapper for a simple MQTT client.
type Client interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Publish(ctx context.Context, topic Topic, message interface{}) error
	// PublishRPC(ctx context.Context, topic Topic, message interface{}) error
	SubscribeWithId(ctx context.Context, topic Topic, handler Handler)
	Subscribe(ctx context.Context, topic Topic, handler Handler) error
	GetId() string
	checks.Check
}

var ErrInvalidVersion = errors.New("invalid MQTT version")

// NewClientFromConfig Creates a new MQTT client from the configuration based on the supported version.
func NewClientFromConfig(obs observability.Observability, cfg Configuration) (Client, error) {
	switch cfg.Version {
	case MqttVersion5:
		return NewV5Client(cfg, obs)
	case MqttVersion3:
		return NewV3Client(cfg, obs)
	default:
		return nil, ErrInvalidVersion
	}
}
