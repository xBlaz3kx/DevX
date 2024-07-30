package rabbit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicBuilder(t *testing.T) {
	topic := NewTopic("service")
	topic.AddWord("word")

	assert.EqualValues(t, "service.word", topic.Build())

	topic.AddWord("word2")

	assert.EqualValues(t, "service.word.word2", topic.Build())
}

func TestTopicWord_String(t *testing.T) {
	word := TopicWord("word")

	assert.Equal(t, "word", word.String())

	word = TopicWord("word2")

	assert.Equal(t, "word2", word.String())
}
