package uuid

import "go.mongodb.org/mongo-driver/bson/primitive"

type ID string

func ConvertID(id ID) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(string(id))
}
