package rabbit

import "time"

type ConsumerOpts struct {
	eventTimeout time.Duration
	routines     int
}

type ConsumerOpt func(*ConsumerOpts)

func newConsumerOptions() ConsumerOpts {
	return ConsumerOpts{
		eventTimeout: time.Second * 20,
		routines:     1,
	}
}

func WithDefaultTimeout(timeout time.Duration) ConsumerOpt {
	return func(c *ConsumerOpts) {
		c.eventTimeout = timeout
	}
}

func WithRoutines(routines int) ConsumerOpt {
	return func(c *ConsumerOpts) {
		c.routines = routines
	}
}
