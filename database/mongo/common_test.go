package mongo

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
