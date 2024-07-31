package rabbit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaders_WithError(t *testing.T) {
	headers := NewHeader().WithError(true).Build()
	assert.Contains(t, headers, HeaderValue{Key: HeaderKeyError, Value: true})

	headers = NewHeader().WithError(false).Build()
	assert.Empty(t, headers)

	headers = NewHeader().WithError(true).WithField("SomeOtherAttribute", "SomeValue").Build()
	assert.Contains(t, headers, HeaderValue{Key: HeaderKeyError, Value: true})
	assert.Contains(t, headers, HeaderValue{Key: "SomeOtherAttribute", Value: "SomeValue"})
}

func TestHeader_WithMethod(t *testing.T) {
	headers := NewHeader().WithError(true).WithField("SomeOtherAttribute", "SomeValue").WithMethod("SomeMethod").Build()
	assert.Contains(t, headers, HeaderValue{Key: HeaderKeyError, Value: true})
	assert.Contains(t, headers, HeaderValue{Key: "SomeOtherAttribute", Value: "SomeValue"})
	assert.Contains(t, headers, HeaderValue{Key: HeaderKeyMethod, Value: "SomeMethod"})

	headers = NewHeader().WithMethod("SomeMethod2").Build()
	assert.Contains(t, headers, HeaderValue{Key: HeaderKeyMethod, Value: "SomeMethod2"})
}

func TestHeader_WithField(t *testing.T) {
	headers := NewHeader().WithField("SomeOtherAttribute", "SomeValue").WithMethod("SomeMethod").Build()
	assert.Contains(t, headers, HeaderValue{Key: "SomeOtherAttribute", Value: "SomeValue"})
	assert.Contains(t, headers, HeaderValue{Key: HeaderKeyMethod, Value: "SomeMethod"})

	headers = NewHeader().WithField("ExampleField", "").Build()
	assert.Contains(t, headers, HeaderValue{Key: "ExampleField", Value: ""})
}
