// nolint:all
package util

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPaginationOptsFromContext(t *testing.T) {
	baseCtx := context.Background()
	ctx := context.WithValue(baseCtx, "limit", "10")
	ctx = context.WithValue(ctx, "offset", "10")

	limit, offset := GetPaginationOptsFromContext(ctx)
	assert.Equal(t, 10, limit)
	assert.Equal(t, 10, offset)

	ctx = context.WithValue(ctx, "limit", "20")
	ctx = context.WithValue(ctx, "offset", "10")

	limit, offset = GetPaginationOptsFromContext(ctx)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 10, offset)

	limit, offset = GetPaginationOptsFromContext(context.Background())
	assert.Equal(t, 30, limit)
	assert.Equal(t, 0, offset)

	limit, offset = GetPaginationOptsFromContext(context.Background(), WithDefaultLimit(100))
	assert.Equal(t, 100, limit)
	assert.Equal(t, 0, offset)
}

func TestGetIntFromContext(t *testing.T) {
	baseCtx := context.Background()
	ctx := context.WithValue(baseCtx, "key", "10")

	val := getIntFromContext(ctx, "key", 0)
	assert.Equal(t, 10, val)

	val = getIntFromContext(ctx, "invalid_key", 0)
	assert.Equal(t, 0, val)

	val = getIntFromContext(nil, "key", 0)
	assert.Equal(t, 0, val)
}
