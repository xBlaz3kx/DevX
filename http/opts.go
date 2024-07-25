package http

import "time"

// Cross Origin Resource Sharing configuration
type CORS struct {
	// IsEnabled is the flag to enable/disable CORS
	IsEnabled bool `yaml:"enabled" json:"enabled,omitempty" mapstructure:"enabled"`

	// AllowedOrigins is the list of allowed origins
	AllowedOrigins []string `yaml:"allowedOrigins" json:"allowedOrigins,omitempty" mapstructure:"allowedOrigins"`
}

type Options struct {
	timeout time.Duration
}

func newOptions() *Options {
	return &Options{
		timeout: 10 * time.Second,
	}
}

func WithTimeout(timeout time.Duration) func(options *Options) {
	return func(options *Options) {
		options.timeout = timeout
	}
}
