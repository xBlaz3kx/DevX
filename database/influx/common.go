package influx

import (
	"context"
	"strconv"

	"github.com/samber/lo"
)

const (
	defaultLimit  = 20
	defaultOffset = 0
	maxLimit      = 100
)

func PaginationFromCtx(ctx context.Context) (int, *int) {
	if ctx == nil {
		return defaultLimit, lo.ToPtr(defaultOffset)
	}

	// Default limit is 20.
	limit := getInt64FromContext(ctx, "limit")
	if limit == nil {
		limit = lo.ToPtr(int64(defaultLimit))
	}

	// Limit should not be more than 100
	if *limit > maxLimit {
		limit = lo.ToPtr(int64(maxLimit))
	}

	// Default offset is 0
	offset := getInt64FromContext(ctx, "offset")
	if offset == nil {
		offset = lo.ToPtr(int64(defaultOffset))
	}

	offsetAsInt := int(*offset)

	return int(*limit), &offsetAsInt
}

// Get int64 value from context
func getInt64FromContext(ctx context.Context, key string) *int64 {
	var val *int64

	variable := ctx.Value(key)
	if variable != nil {

		parseInt, err := strconv.ParseInt(variable.(string), 10, 64)
		if err != nil {
			return nil
		}

		val = &parseInt
	}

	return val
}
