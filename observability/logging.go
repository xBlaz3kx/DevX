package observability

import (
	"os"

	"github.com/GLCharge/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

	logger := otelzap.New(
		zap.New(core),
		otelzap.WithTraceIDField(true),
		otelzap.WithMinLevel(level),
	)

	return &Logging{
		logger: logger,
	}
}
