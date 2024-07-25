package observability

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
		Enabled bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
		Address string `json:"address,omitempty" yaml:"address" mapstructure:"address"`
		TLS     TLS    `json:"tls" yaml:"tls" mapstructure:"tls"`
	}

	// MetricsConfig configures the metrics for the application (over OpenTelemetry GRPC).
	MetricsConfig struct {
		Enabled      bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
		Address      string `json:"address,omitempty" yaml:"address" mapstructure:"address"`
		TLS          TLS    `json:"tls" yaml:"tls" mapstructure:"tls"`
		PushInterval int64  `json:"pushInterval,omitempty" yaml:"pushInterval" mapstructure:"pushInterval"`
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

	// TLS configuration with the option to enable/disable and with paths to the certificates
	TLS struct {
		// IsEnabled is the flag to enable/disable TLS
		IsEnabled bool `yaml:"enabled" json:"enabled,omitempty" mapstructure:"enabled"`

		// RootCertificatePath is the path to the root certificate
		RootCertificatePath string `yaml:"rootCaPath" json:"rootCaPath,omitempty" mapstructure:"rootCaPath"`

		// CertificatePath is the path to the certificate
		CertificatePath string `yaml:"certPath" json:"certPath,omitempty" mapstructure:"certPath"`

		// PrivateKeyPath is the path to the private key
		PrivateKeyPath string `yaml:"keyPath" json:"keyPath,omitempty" mapstructure:"keyPath"`
	}
)
