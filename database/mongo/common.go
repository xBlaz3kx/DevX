package mongo

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvalidMongoId = errors.New("invalid mongo id")

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
