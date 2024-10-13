package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	mqtt "github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/xBlaz3kx/DevX/observability"
	"go.uber.org/zap"
)

type (
	HandlerV5 func(client ClientV5, topicIds []string, payloadId uint16, payload interface{}, err error)

	// ClientV5 is an interface wrapper for a simple MQTT client.
	ClientV5 interface {
		Connect(ctx context.Context)
		Disconnect()
		Publish(ctx context.Context, topic Topic, message interface{}) error
		SubscribeWithId(ctx context.Context, topic Topic, handler HandlerV5)
		Subscribe(ctx context.Context, topic Topic, handler HandlerV5)
	}

	// V5Impl concrete implementation of the ClientV5, which is essentially a wrapper over the mqtt lib.
	V5Impl struct {
		mqttClient *mqtt.ConnectionManager
		brokerUrl  *url.URL
		clientId   string
		obs        observability.Observability
	}
)

// NewMqttV5Client creates a wrapped mqtt ClientV5 with specific settings.
func NewMqttV5Client(clientSettings Configuration, obs observability.Observability) ClientV5 {
	obs.Log().Info("Creating a new MQTT client..")
	broker := fmt.Sprintf("tcp://%s", clientSettings.Address)
	clientId, _ := os.Hostname()
	parse, err := url.Parse(broker)
	if err != nil {
		return nil
	}

	return &V5Impl{
		mqttClient: nil,
		brokerUrl:  parse,
		clientId:   clientId,
		obs:        obs,
	}
}

func (c *V5Impl) Connect(ctx context.Context) {
	c.obs.Log().Debug("Connecting to the broker")

	cliCfg := mqtt.ClientConfig{
		BrokerUrls:        []*url.URL{c.brokerUrl},
		KeepAlive:         10,
		ConnectRetryDelay: 3,
		OnConnectionUp: func(cm *mqtt.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("mqtt connected to broker")
		},
		OnConnectError: func(err error) { fmt.Printf("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID:      c.clientId,
			OnClientError: func(err error) { fmt.Printf("server requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cm, err := mqtt.NewConnection(ctx, cliCfg)
	if err != nil {
		panic(err)
	}

	c.mqttClient = cm
}

func (c *V5Impl) Disconnect() {
	c.obs.Log().Debug("Disconnecting the MQTT client")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := c.mqttClient.Disconnect(ctx)
	if err != nil {
		return
	}
}

// Publish a new message to a topic
func (c *V5Impl) Publish(ctx context.Context, topic Topic, message interface{}) error {
	logInfo := c.obs.Log().With(
		zap.String("topic", string(topic)),
		zap.Any("message", message),
	)
	logInfo.Debug("Publishing a message to topic")

	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
func (c *V5Impl) Subscribe(ctx context.Context, topic Topic, handler HandlerV5) {
	logInfo := c.obs.Log().With(
		zap.String("topic", string(topic)),
	)
	logInfo.Debug("Subscribing to a topic")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	subscribe := paho.Subscribe{
		Properties: &paho.SubscribeProperties{
			SubscriptionIdentifier: nil,
			User:                   nil,
		},
		Subscriptions: nil,
	}

	_, _ = c.mqttClient.Subscribe(ctx, &subscribe)
}

// SubscribeWithId to a topic
func (c *V5Impl) SubscribeWithId(ctx context.Context, topic Topic, handler HandlerV5) {
	logInfo := c.obs.Log().With(
		zap.String("topic", string(topic)),
	)
	logInfo.Debug("Subscribing to a topic")
}
