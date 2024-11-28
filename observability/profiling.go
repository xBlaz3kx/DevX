package observability

import (
	"github.com/GLCharge/otelzap"
	"github.com/grafana/pyroscope-go"
	"github.com/pkg/errors"
)

type ProfilingConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Address string `json:"address,omitempty" yaml:"address" mapstructure:"address"`
	// Deprecated: Use username and password instead
	AuthToken string `json:"authToken,omitempty" yaml:"authToken" mapstructure:"authToken"`
	Username  string `json:"username,omitempty" yaml:"username" mapstructure:"username"`
	Password  string `json:"password,omitempty" yaml:"password" mapstructure:"password"`
}

type Profiling struct {
	profiler *pyroscope.Profiler
}

func NewProfiler(name string, logger *otelzap.Logger, config ProfilingConfig) (*Profiling, error) {
	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName:   name,
		ServerAddress:     config.Address,
		AuthToken:         config.AuthToken,
		BasicAuthUser:     config.Username,
		BasicAuthPassword: config.Password,
		Logger:            logger.Sugar(),
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to start profiler")
	}

	return &Profiling{
		profiler: profiler,
	}, nil
}

func (p *Profiling) Shutdown() error {
	if p == nil {
		return nil
	}

	// Flush before stopping
	p.profiler.Flush(false)

	return p.profiler.Stop()
}
