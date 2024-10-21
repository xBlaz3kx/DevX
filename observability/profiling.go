package observability

import (
	"github.com/grafana/pyroscope-go"
	"github.com/pkg/errors"
)

type ProfilingConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Address   string `json:"address,omitempty" yaml:"address" mapstructure:"address"`
	AuthToken string `json:"authToken,omitempty" yaml:"authToken" mapstructure:"authToken"`
}

type Profiling struct {
	profiler *pyroscope.Profiler
}

func NewProfiler(name string, config ProfilingConfig) (*Profiling, error) {
	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: name,
		ServerAddress:   config.Address,
		AuthToken:       config.AuthToken,
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

	return p.profiler.Stop()
}
