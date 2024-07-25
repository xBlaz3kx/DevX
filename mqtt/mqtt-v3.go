package mqtt

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/xBlaz3kx/DevX/observability"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

type (
	Topic string

	MessageHandler func(client Client, topicIds []string, payloadId uint16, payload interface{}, err error)

	// Client is an interface wrapper for a simple MQTT client.
	Client interface {
		Connect()
		Disconnect()
		Publish(topic Topic, message interface{}) error
		Subscribe(topic Topic, handler MessageHandler)
		SubscribeToAny(topic Topic, handler MessageHandler)
		GetId() string
	}

	// clientImpl concrete implementation of the Client, which is essentially a wrapper over the mqtt lib.
	clientImpl struct {
		mqttClient mqtt.Client
		id         string
		obs        observability.Observability
	}
)

func (t Topic) String() string {
	return string(t)
}

// NewMqttClient creates a wrapped mqtt Client with specific settings.
func NewMqttClient(clientSettings Configuration, obs observability.Observability) Client {
	obs.Log().Info("Creating a new MQTT client ...")
	broker := fmt.Sprintf("tcps://%s", clientSettings.Address)

	// Basic client settings
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientSettings.ClientId)
	opts.SetUsername(clientSettings.Username)
	opts.SetPassword(clientSettings.Password)

	// Connection settings
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetMaxReconnectInterval(time.Second * 5)

	// Set TLS settings (required by AWS)
	opts.SetTLSConfig(createTlsConfiguration(obs, clientSettings.TLS))

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		obs.Log().Info("Connected to broker")
	})

	opts.SetDefaultPublishHandler(func(client mqtt.Client, message mqtt.Message) {
		obs.Log().Sugar().Infof("Received message %s from topic %s", message.Payload(), message.Topic())
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		obs.Log().Info("Disconnected from broker", zap.Error(err))
	})

	// Connect to the MQTT broker
	client := mqtt.NewClient(opts)
	return &clientImpl{
		mqttClient: client,
		id:         clientSettings.ClientId,
		obs:        obs.WithSpanKind(trace.SpanKindClient),
	}
}

func (c *clientImpl) Connect() {
	c.obs.Log().Debug("Connecting to the MQTT broker")
	c.mqttClient.Connect().Wait()
}

func (c *clientImpl) Disconnect() {
	c.obs.Log().Debug("Disconnecting the MQTT client")
	c.mqttClient.Disconnect(100)
}

func (c *clientImpl) GetId() string {
	return c.id
}

// Publish a new message to a topic
func (c *clientImpl) Publish(topic Topic, message interface{}) error {
	logInfo := c.obs.Log().With(
		zap.String("topic", string(topic)),
		zap.Any("message", message),
	)
	logInfo.Debug("Publishing a message to topic")

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	token := c.mqttClient.Publish(topic.String(), 1, false, jsonMessage)
	go func(token mqtt.Token) {
		if token.Error() != nil {
			c.obs.Log().Warn("Token error", zap.Error(token.Error()))
		}
	}(token)
	return nil
}

// SubscribeToAny to a topic
func (c *clientImpl) SubscribeToAny(topic Topic, handler MessageHandler) {
	logInfo := c.obs.Log().With(zap.String("topic", string(topic)))
	logInfo.Debug("Subscribing to a topic")

	token := c.mqttClient.Subscribe(topic.String(), 1, func(client mqtt.Client, message mqtt.Message) {
		var data interface{}

		// Transform the payload to the object and pass it to the handler function for further processing
		err := json.Unmarshal(message.Payload(), &data)
		if err != nil {
			logInfo.Sugar().Errorf("Error parsing the data: %v", err)
			return
		}

		// Parse the topic and get the Ids based on the original topic.
		ids, err := GetIdsFromTopic(c.obs, message.Topic(), topic)
		if err != nil {
			logInfo.Sugar().Errorf("Error getting the topic info: %v", err)
			return
		}

		handler(c, ids, message.MessageID(), data, err)
	})

	go func(token mqtt.Token) {
		token.Wait()
		if token.Error() != nil {
			logInfo.Warn("Token error", zap.Error(token.Error()))
		}
	}(token)
}

// Subscribe to a topic
func (c *clientImpl) Subscribe(topic Topic, handler MessageHandler) {
	logInfo := c.obs.Log().With(zap.String("topic", string(topic)))
	logInfo.Debug("Subscribing to a topic")

	token := c.mqttClient.Subscribe(topic.String(), 1, func(client mqtt.Client, message mqtt.Message) {
		var data interface{}

		// Transform the payload to the object and pass it to the handler function for further processing
		err := json.Unmarshal(message.Payload(), &data)
		if err != nil {
			logInfo.Sugar().Errorf("Error parsing the data: %v", err)
			return
		}

		ids := strings.Split(message.Topic(), "/")
		handler(c, ids, message.MessageID(), data, err)
	})

	go func(token mqtt.Token) {
		token.Wait()
		if token.Error() != nil {
			c.obs.Log().Warn("Token error", zap.Error(token.Error()))
		}
	}(token)
}
