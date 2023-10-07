package activities

import (
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/mongo"
)

type Activities struct {
	*mongo.Collection
	*service.MongoDBService[Activity]
}

func NewActivities(c *mongo.Collection) *Activities {
	return &Activities{c, service.New[Activity](c)}
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
	ID          service.ID    `json:"id" bson:"_id"`
	UserID      service.ID    `json:"userID" bson:"userID"`
	CreatedDate *service.Time `json:"createdDate" bson:"createdDate"`
	Type        ActivityType  `json:"type" bson:"type"`
	Value       float64       `json:"value" bson:"value"`
	Start       time.Time     `json:"start" bson:"start"`
	End         time.Time     `json:"end,omitempty" bson:"end"`
}

func (a Activity) GetID() service.ID {
	return a.ID
}

func (a Activity) GetCreatedDate() service.Time {
	return *a.CreatedDate
}

func (a Activity) CanBeReadBy(userID service.ID, admin bool) bool {
	return a.UserID == userID || admin
}

func (a Activity) CanBeUpdatedBy(userID service.ID, admin bool) bool {
	return a.UserID == userID || admin
}

func (a Activity) CanBeDeletedBy(userID service.ID, admin bool) bool {
	return a.UserID == userID || admin
}

func New(Type ActivityType, Value float64) Activity {
	return Activity{
		ID:    service.NewID(),
		Type:  Type,
		Value: Value,
	}
}

var (
	_ service.Resource              = (*Activity)(nil)
	_ service.CRUDService[Activity] = (*Activities)(nil)
)
