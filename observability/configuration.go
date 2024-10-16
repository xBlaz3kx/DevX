package observability

import "github.com/xBlaz3kx/DevX/tls"

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type (
	// LogConfig is the configuration for the logging
	LogConfig struct {
		Level *LogLevel `yaml:"level" json:"level,omitempty" mapstructure:"level"` // debug, info, warn, error
	}

	LogLevel string

	// Config configures logs, traces and metrics for the application
	Config struct {
		Tracing   TracingConfig   `json:"tracing" yaml:"tracing" mapstructure:"tracing"`
		Metrics   MetricsConfig   `json:"metrics" yaml:"metrics" mapstructure:"metrics"`
		Logging   LogConfig       `json:"logging" yaml:"logging" mapstructure:"logging"`
		Profiling ProfilingConfig `json:"profiling" yaml:"profiling" mapstructure:"profiling"`
	}

	// TracingConfig configures the tracing for the application (over OpenTelemetry GRPC).
	TracingConfig struct {
		Enabled bool    `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
		Address string  `json:"address,omitempty" yaml:"address" mapstructure:"address"`
		TLS     tls.TLS `json:"tls" yaml:"tls" mapstructure:"tls"`
	}

	// MetricsConfig configures the metrics for the application (over OpenTelemetry GRPC).
	MetricsConfig struct {
		Enabled      bool    `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
		Address      string  `json:"address,omitempty" yaml:"address" mapstructure:"address"`
		TLS          tls.TLS `json:"tls" yaml:"tls" mapstructure:"tls"`
		PushInterval int64   `json:"pushInterval,omitempty" yaml:"pushInterval" mapstructure:"pushInterval"`
	}

	ProfilingConfig struct {
		Enabled   bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
		Address   string `json:"address,omitempty" yaml:"address" mapstructure:"address"`
		AuthToken string `json:"authToken,omitempty" yaml:"authToken" mapstructure:"authToken"`
	}

	ServiceInfo struct {
		Name    string
		Version string
	}
)
