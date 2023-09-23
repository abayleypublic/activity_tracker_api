package activities

import (
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
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

// TODO - initialise ID via a New function
type Activity struct {
	ID    uuid.ID      `json:"id" bson:"_id"`
	Type  ActivityType `json:"name" bson:"name"`
	Value float64      `json:"value" bson:"value"`
}
