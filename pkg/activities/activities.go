package activities

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Activities struct {
	*mongo.Collection
}

func NewActivities(c *mongo.Collection) *Activities {
	return &Activities{c}
}

type ActivityName string

const (
	Running ActivityName = "running"
)

type Activity struct {
	Name ActivityName `json:"name" bson:"name"`
}
