package observability

import (
	"context"
	"os"

	"github.com/GLCharge/otelzap"
	"github.com/spf13/viper"
	"github.com/xBlaz3kx/DevX/tls"
	oZap "go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otelLog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogFormat string

// LogConfig is the configuration for the logging
type LogConfig struct {
	Level      *LogLevel   `yaml:"level" json:"level,omitempty" mapstructure:"level"` // debug, info, warn, error
	OtelLogger *OtelLogger `yaml:"otelLogger" json:"otelLogger,omitempty" mapstructure:"otelLogger"`
}

type OtelLogger struct {
	Address string   `yaml:"address" json:"address,omitempty" mapstructure:"address"`
	TLS     *tls.TLS `yaml:"tls" json:"tls,omitempty" mapstructure:"tls"`
}

type LogLevel string

type Logging struct {
	logger *otelzap.Logger
}

func (l *Logging) Logger() *otelzap.Logger {
	if l == nil || l.logger == nil {
		logger, _ := zap.NewProduction()
		l.logger = otelzap.New(logger)
	}

	return l.logger.Clone()
}

func (l *Logging) Shutdown() error {
	if l == nil {
		return nil
	}

	return l.logger.Sync()
}

func NewLogging(config LogConfig) *Logging {
	// Default to info level
	level := zapcore.InfoLevel
	if config.Level != nil {
		switch *config.Level {
		case LogLevelDebug:
			level = zapcore.DebugLevel
		case LogLevelWarn:
			level = zapcore.WarnLevel
		case LogLevelError:
			level = zapcore.ErrorLevel
		}
	}

	stdout := zapcore.Lock(os.Stdout)
	stderr := zapcore.Lock(os.Stderr)

	stdoutLevelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= level && l < zapcore.ErrorLevel
	})
	stderrLevelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= level && l >= zapcore.ErrorLevel
	})

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, stdout, stdoutLevelEnabler),
		zapcore.NewCore(encoder, stderr, stderrLevelEnabler),
	)

	if config.OtelLogger != nil {
		exporter, err := otlploggrpc.New(context.Background(),
			otlploggrpc.WithEndpoint(config.OtelLogger.Address),
			otlploggrpc.WithInsecure(),
		)
		if err != nil {
			return nil
		}

		// Setup Otel logger
		env := viper.GetString("environment")

		var processor otelLog.Processor
		if env == "production" {
			processor = otelLog.NewSimpleProcessor(exporter)
		} else {
			processor = otelLog.NewBatchProcessor(exporter)
		}

		logProvider := otelLog.NewLoggerProvider(otelLog.WithProcessor(processor))

		// Override the core
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, stdout, stdoutLevelEnabler),
			zapcore.NewCore(encoder, stderr, stderrLevelEnabler),
			oZap.NewCore("", oZap.WithLoggerProvider(logProvider)),
		)
	}

	logger := otelzap.New(
		zap.New(core),
		otelzap.WithTraceIDField(true),
		otelzap.WithMinLevel(level),
	)

	return &Logging{
		logger: logger,
	}
}
