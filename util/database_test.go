// nolint:all
package util

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
}

func TestStringToObjectId(t *testing.T) {
	// Test valid input
	validInput := []string{
		"5f6b9e9e56a0f0a3c8d17a91",
		"5f6b9e9e56a0f0a3c8d17a92",
		"5f6b9e9e56a0f0a3c8d17a93",
	}
	expected := []primitive.ObjectID{
		{0x5f, 0x6b, 0x9e, 0x9e, 0x56, 0xa0, 0xf0, 0xa3, 0xc8, 0xd1, 0x7a, 0x91},
		{0x5f, 0x6b, 0x9e, 0x9e, 0x56, 0xa0, 0xf0, 0xa3, 0xc8, 0xd1, 0x7a, 0x92},
		{0x5f, 0x6b, 0x9e, 0x9e, 0x56, 0xa0, 0xf0, 0xa3, 0xc8, 0xd1, 0x7a, 0x93},
	}

	result, _ := StringToObjectId(validInput...)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Test invalid input
	invalidInput := []string{
		"invalid_id",
		"also_invalid",
	}
	result, _ = StringToObjectId(invalidInput...)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}
