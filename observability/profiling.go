package observability

import (
	"github.com/grafana/pyroscope-go"
	"github.com/pkg/errors"
)

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

func (p *Profiling) Shutdown() {
	if p == nil {
		return
	}
	_ = p.profiler.Stop()
}
