package util

import (
	"context"
	"strconv"
)

// GetPaginationOptsFromContext gets the limit and offset from the context
func GetPaginationOptsFromContext(ctx context.Context) (int, int) {
	return getIntFromContext(ctx, "limit", 30), getIntFromContext(ctx, "offset", 0)
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
