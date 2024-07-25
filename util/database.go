package util

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvalidMongoId = errors.New("invalid mongo id")

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

// StringToObjectId converts a string to a primitive.ObjectID
func StringToObjectId(ids ...string) ([]primitive.ObjectID, error) {
	objectIds := []primitive.ObjectID{}

	for _, id := range ids {
		hex, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, ErrInvalidMongoId
		}

		objectIds = append(objectIds, hex)
	}

	return objectIds, nil
}
