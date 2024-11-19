package influx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationFromCtx(t *testing.T) {
	cases := []struct {
		name   string
		ctx    context.Context
		limit  int
		offset int
	}{
		{
			name:   "Nil context",
			ctx:    nil,
			limit:  defaultLimit,
			offset: defaultOffset,
		},
		{
			name:   "Limit and offset in context",
			ctx:    context.WithValue(context.WithValue(context.Background(), "limit", "10"), "offset", "3"),
			limit:  10,
			offset: 3,
		},
		{
			name:   "Limit in context",
			ctx:    context.WithValue(context.Background(), "limit", "10"),
			limit:  10,
			offset: defaultOffset,
		},
		{
			name:   "Offset in context",
			ctx:    context.WithValue(context.Background(), "offset", "10"),
			limit:  defaultLimit,
			offset: 10,
		},
		{
			name:   "Limit more than 100",
			ctx:    context.WithValue(context.Background(), "limit", "101"),
			limit:  maxLimit,
			offset: defaultOffset,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset := PaginationFromCtx(tt.ctx)
			assert.EqualValues(t, tt.limit, limit)
			assert.EqualValues(t, tt.offset, *offset)
		})
	}
}
