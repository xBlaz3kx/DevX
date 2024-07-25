package rabbit

type exchangeType string

const (
	TopicExchange  exchangeType = "topic"
	DirectExchange exchangeType = "direct"
	FanoutExchange exchangeType = "fanout"
)

type Exchange string

func (t Exchange) String() string {
	return string(t)
}

// Exchanges
const (
	CentralExchange Exchange = "CENTRAL"

	GlobalNotificationExchange Exchange = "NOTIFICATION"
)
