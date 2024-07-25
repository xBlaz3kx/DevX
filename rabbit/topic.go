package rabbit

import "fmt"

type TopicWord string

func (t TopicWord) String() string {
	return string(t)
}

type Topic string

func (t Topic) String() string {
	return string(t)
}

const (
	replyBase TopicWord = "REPLY"
)

type TopicBuilder struct {
	topic Topic
}

// NewTopic creates a new topic that starts with the exchange of the receiving service
func NewTopic(service Exchange) *TopicBuilder {
	return &TopicBuilder{topic: Topic(service)}
}

// AddWord adds a new word to the topic
func (rh *TopicBuilder) AddWord(word TopicWord) *TopicBuilder {
	rh.topic = Topic(fmt.Sprintf("%s.%s", rh.topic, word))
	return rh
}

// Build returns a topic string
func (rh *TopicBuilder) Build() Topic {
	return rh.topic
}
