package mqtt

import (
	"context"
	"encoding/json"
	"net/url"

	mqtt "github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/xBlaz3kx/DevX/observability"
	"go.uber.org/zap"
)

// mqttV5 concrete implementation of the ClientV5, which is essentially a wrapper over the mqtt lib.
type mqttV5 struct {
	mqttClient *mqtt.ConnectionManager
	brokerUrl  *url.URL
	clientId   string
	obs        observability.Observability
}

// NewV5Client creates a wrapped mqtt ClientV5 with specific settings.
func NewV5Client(clientSettings Configuration, obs observability.Observability) (Client, error) {
	obs.Log().Info("Creating a new MQTT client..")

	parse, err := url.Parse(clientSettings.Address)
	if err != nil {
		return nil, err
	}

	return &mqttV5{
		mqttClient: nil,
		brokerUrl:  parse,
		clientId:   clientSettings.ClientId,
		obs:        obs,
	}, nil
}

func (c *mqttV5) Connect(ctx context.Context) error {
	c.obs.Log().Debug("Connecting to the broker")

	clientConfig := mqtt.ClientConfig{
		BrokerUrls: []*url.URL{c.brokerUrl},
		KeepAlive:  10,
		OnConnectionUp: func(cm *mqtt.ConnectionManager, connAck *paho.Connack) {
			c.obs.Log().Debug("Client connected to broker")
		},
		OnConnectError: func(err error) { c.obs.Log().With(zap.Error(err)).Debug("error whilst attempting connection") },
		ClientConfig: paho.ClientConfig{
			ClientID:      c.clientId,
			OnClientError: func(err error) { c.obs.Log().With(zap.Error(err)).Error("server requested disconnect") },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					c.obs.Log().With(zap.String("reason", d.Properties.ReasonString)).Error("server requested disconnect: %s\n")
				}
			},
		},
	}

	cm, err := mqtt.NewConnection(ctx, clientConfig)
	if err != nil {
		return err
	}

	c.mqttClient = cm
	return nil
}

func (c *mqttV5) Disconnect(ctx context.Context) error {
	c.obs.Log().Debug("Disconnecting the MQTT client")

	err := c.mqttClient.Disconnect(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Publish a new message to a topic
func (c *mqttV5) Publish(ctx context.Context, topic Topic, message interface{}) error {
	logInfo := c.obs.Log().With(
		zap.String("topic", string(topic)),
		zap.Any("message", message),
	)
	logInfo.Debug("Publishing a message to topic")

	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	publishMessage := paho.Publish{
		QoS:     0,
		Retain:  false,
		Topic:   topic.String(),
		Payload: payload,
		Properties: &paho.PublishProperties{
			CorrelationData:        nil,
			ContentType:            "application/json",
			ResponseTopic:          "",
			PayloadFormat:          nil,
			MessageExpiry:          nil,
			SubscriptionIdentifier: nil,
			TopicAlias:             nil,
			User:                   nil,
		},
	}

	_, err = c.mqttClient.Publish(ctx, &publishMessage)
	return err
}

// Subscribe to a topic
func (c *mqttV5) Subscribe(ctx context.Context, topic Topic, handler Handler) error {
	logInfo := c.obs.Log().With(
		zap.String("topic", topic.String()),
	)
	logInfo.Debug("Subscribing to a topic")

	subscribe := paho.Subscribe{
		Properties: &paho.SubscribeProperties{
			SubscriptionIdentifier: nil,
			User:                   nil,
		},
		Subscriptions: nil,
	}

	_, err := c.mqttClient.Subscribe(ctx, &subscribe)
	if err != nil {
		return err
	}

	return nil
}

// SubscribeWithId to a topic
func (c *mqttV5) SubscribeWithId(ctx context.Context, topic Topic, handler Handler) {
	logInfo := c.obs.Log().With(
		zap.String("topic", string(topic)),
	)
	logInfo.Debug("Subscribing to a topic")
}

func (c *mqttV5) GetId() string {
	return c.clientId
}

func (c *mqttV5) Pass() bool {
	return true
}

func (c *mqttV5) Name() string {
	return "mqtt-v5"
}
