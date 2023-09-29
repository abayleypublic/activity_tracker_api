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

type ActivityCategory map[ActivityType]struct{}

var (
	Moving = ActivityCategory{
		Walking:  {},
		Running:  {},
		Swimming: {},
		Cycling:  {},
	}
)

type Activity struct {
	ID    uuid.ID      `json:"id" bson:"_id"`
	Type  ActivityType `json:"type" bson:"type"`
	Value float64      `json:"value" bson:"value"`
}

func New(Type ActivityType, Value float64) Activity {
	return Activity{
		ID:    uuid.NewID(),
		Type:  Type,
		Value: Value,
	}
}
