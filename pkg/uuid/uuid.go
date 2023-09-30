package uuid

import (
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ID is a type alias for string, intended to represent a UUID.
type ID string

// ConvertID converts a UUID of type ID to a MongoDB ObjectID.
// It returns the converted ObjectID and any error encountered during the conversion.
func ConvertID(id ID) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(string(id))
}

func NewID() ID {
	hex := primitive.NewObjectID().Hex()
	log.Println(hex)
	return ID(hex)
}
