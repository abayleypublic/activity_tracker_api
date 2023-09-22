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

type ActivityType string

const (
	Any      ActivityType = "any"
	Walking  ActivityType = "walking"
	Running  ActivityType = "running"
	Swimming ActivityType = "swimming"
	Cycling  ActivityType = "cycling"
)

var (
	Moving map[ActivityType]struct{} = map[ActivityType]struct{}{
		Walking:  {},
		Running:  {},
		Swimming: {},
		Cycling:  {},
	}
)

type Activity struct {
	Type  ActivityType `json:"name" bson:"name"`
	Value float64      `json:"value" bson:"value"`
}
