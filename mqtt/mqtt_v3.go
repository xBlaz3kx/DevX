package mqtt

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/xBlaz3kx/DevX/observability"
	"go.uber.org/zap"
)

// mqttV3 concrete implementation of the Client, which is essentially a wrapper over the mqtt lib.
type mqttV3 struct {
	obs        observability.Observability
	mqttClient mqtt.Client
	id         string
}

// NewV3Client creates a wrapped mqtt Client with specific settings.
func NewV3Client(clientSettings Configuration, obs observability.Observability) (Client, error) {
	// Basic client settings
	opts := mqtt.NewClientOptions()
	opts.AddBroker(clientSettings.Address)
	opts.SetClientID(clientSettings.ClientId)
	opts.SetUsername(clientSettings.Username)
	opts.SetPassword(clientSettings.Password)

	// Connection settings
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetMaxReconnectInterval(time.Second * 5)

	// Append certs if enabled
	if clientSettings.TLS.IsEnabled {
		tlsSettings, err := clientSettings.TLS.ToTlsConfig()
		if err != nil {
			return nil, err
		}

		opts.SetTLSConfig(tlsSettings)
	}

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
	return &mqttV3{
		mqttClient: client,
		id:         clientSettings.ClientId,
		obs:        obs,
	}, nil
}

func (c *mqttV3) Connect(_ context.Context) error {
	c.obs.Log().Debug("Connecting to the MQTT broker")
	c.mqttClient.Connect().Wait()
	return nil
}

func (c *mqttV3) Disconnect(_ context.Context) error {
	c.obs.Log().Debug("Disconnecting the MQTT client")
	c.mqttClient.Disconnect(100)
	return nil
}

func (c *mqttV3) GetId() string {
	return c.id
}

// Publish a new message to a topic
func (c *mqttV3) Publish(_ context.Context, topic Topic, message interface{}) error {
	logInfo := c.obs.Log().With(
		zap.String("topic", topic.String()),
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

// SubscribeWithId to a topic
func (c *mqttV3) SubscribeWithId(_ context.Context, topic Topic, handler Handler) {
	logInfo := c.obs.Log().With(zap.String("topic", string(topic)))
	logInfo.Debug("Subscribing to a topic")

	token := c.mqttClient.Subscribe(topic.String(), 1, func(client mqtt.Client, message mqtt.Message) {
		var data interface{}

		// todo support other types of messages
		// Transform the payload to the object and pass it to the handler function for further processing
		err := json.Unmarshal(message.Payload(), &data)
		if err != nil {
			logInfo.Sugar().Errorf("Error parsing the data: %v", err)

			// Invoke handler with error
			handler(c, nil, message.MessageID(), nil, err)
			return
		}

		// Parse the topic and get the Ids based on the original topic.
		ids, err := GetIdsFromTopic(c.obs.Log(), message.Topic(), topic)
		if err != nil {
			logInfo.Sugar().Errorf("Error getting the topic info: %v", err)

			// Invoke handler with error
			handler(c, nil, message.MessageID(), nil, err)
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
func (c *mqttV3) Subscribe(_ context.Context, topic Topic, handler Handler) error {
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
	return nil
}

func (c *mqttV3) Pass() bool {
	return c.mqttClient.IsConnected()
}

func (c *mqttV3) Name() string {
	return "mqtt-client-v3"
}
