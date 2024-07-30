package rabbit

type HeaderKey string

const (
	HeaderKeyError     HeaderKey = "error"
	HeaderKeyMethod    HeaderKey = "method"
	HeaderKeyReplyType HeaderKey = "reply_type"
)

type HeaderReplyType string

const (
	HeaderReplyTypeAcknowledge        HeaderReplyType = "acknowledge"
	HeaderReplyTypeInterceptedMessage HeaderReplyType = "intercepted_message"
)

type HeaderValue struct {
	Key   HeaderKey
	Value any
}

type Header struct {
	header []HeaderValue
}

func NewHeader() *Header {
	return &Header{header: []HeaderValue{}}
}

func (rh *Header) WithError(err bool) *Header {
	if err == true {
		rh.header = append(rh.header, HeaderValue{Key: HeaderKeyError, Value: err})
	}
	return rh
}

func (rh *Header) WithMethod(method TopicWord) *Header {
	rh.header = append(rh.header, HeaderValue{Key: HeaderKeyMethod, Value: string(method)})
	return rh
}

// When adding custom fields, make sure the value is supported: https://pkg.go.dev/github.com/wagslane/go-rabbitmq#Table
func (rh *Header) WithField(key HeaderKey, value any) *Header {
	rh.header = append(rh.header, HeaderValue{Key: key, Value: value})
	return rh
}

func (rh *Header) Build() []HeaderValue {
	return rh.header
}
