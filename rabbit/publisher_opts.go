package rabbit

type PublishOpt func(*PublisherOptions)

type PublisherOptions struct {
	headers []HeaderValue
	tracing bool
}

func newPublisherOptions() *PublisherOptions {
	return &PublisherOptions{
		headers: make([]HeaderValue, 0),
	}
}

func WithPublisherHeader(header []HeaderValue) func(options *PublisherOptions) {
	return func(options *PublisherOptions) {
		options.headers = append(options.headers, header...)
	}
}

func WithPublisherTracing(tracing bool) func(options *PublisherOptions) {
	return func(options *PublisherOptions) {
		options.tracing = tracing
	}
}
