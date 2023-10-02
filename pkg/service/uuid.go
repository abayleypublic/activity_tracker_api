package service

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ID is a type alias for string, intended to represent a UUID.
type ID string

// ConvertID converts a UUID of type ID to a MongoDB ObjectID.
// It returns the converted ObjectID and any error encountered during the conversion.
func (id ID) ConvertID() string {
	return string(id)
}

func NewID() ID {
	return ID(primitive.NewObjectID().Hex())
}

func (id ID) GetID() ID {
	return id
}

var (
	_ Attribute = (*ID)(nil)
)
