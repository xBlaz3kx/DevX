package util

import (
	"context"
	"strconv"
)

type PaginationOpts struct {
	defaultLimit int
}

type PaginationOpt func(*PaginationOpts)

// WithDefaultLimit sets the default limit for pagination
func WithDefaultLimit(defaultLimit int) PaginationOpt {
	return func(opts *PaginationOpts) {
		opts.defaultLimit = defaultLimit
	}
}

// GetPaginationOptsFromContext gets the limit and offset from the context
func GetPaginationOptsFromContext(ctx context.Context, opts ...PaginationOpt) (int, int) {
	paginationOpts := &PaginationOpts{
		defaultLimit: 30,
	}
	for _, opt := range opts {
		opt(paginationOpts)
	}

	return getIntFromContext(ctx, "limit", paginationOpts.defaultLimit), getIntFromContext(ctx, "offset", 0)
}

// GetIntFromContext returns the int value from the context
func getIntFromContext(ctx context.Context, key string, defaultNum int) int {
	if ctx == nil {
		return defaultNum
	}

	variable := ctx.Value(key)
	if variable != nil {
		parseInt, err := strconv.ParseInt(variable.(string), 10, 32)
		if err != nil {
			return defaultNum
		}

		return int(parseInt)
	}

	return defaultNum
}
