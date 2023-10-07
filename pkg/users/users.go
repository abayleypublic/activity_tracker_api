// Package users provides the structure and methods for handling user data in the application.
package users

import (
	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Users is a wrapper around a MongoDB collection of users.
type Users struct {
	*mongo.Collection
	*service.MongoDBService[User]
	activities *activities.Activities
}

// NewUsers creates a new Users instance with the provided MongoDB collection.
func NewUsers(c *mongo.Collection, a *activities.Activities) *Users {
	return &Users{c, service.New[User](c), a}
}

// PartialUser represents a user with only the ID, first name, and last name fields.
type PartialUser struct {
	ID        service.ID `json:"id,omitempty" bson:"_id,omitempty"`
	FirstName string     `json:"firstName,omitempty" bson:"firstName"`
	LastName  string     `json:"lastName,omitempty" bson:"lastName"`
}

func (u PartialUser) GetID() service.ID {
	return u.ID
}

// User represents a full user with all fields, including activities.
type User struct {
	PartialUser `bson:",inline"`
	CreatedDate *service.Time `json:"createdDate" bson:"createdDate"`
	Email       string        `json:"email,omitempty" bson:"email"`
	Bio         string        `json:"bio,omitempty" bson:"bio"`
	Challenges  []service.ID  `json:"challenges" bson:"challenges"`
}

func (u User) GetCreatedDate() service.Time {
	return *u.CreatedDate
}

func (u User) CanBeReadBy(userID service.ID, admin bool) bool {
	return u.ID == userID || admin
}

func (u User) CanBeUpdatedBy(userID service.ID, admin bool) bool {
	return u.ID == userID || admin
}

func (u User) CanBeDeletedBy(userID service.ID, admin bool) bool {
	return u.ID == userID || admin
}

func (u *User) MarshalBSON() ([]byte, error) {
	type RawUser User

	if u.Challenges == nil {
		u.Challenges = make([]service.ID, 0)
	}

	return bson.Marshal((*RawUser)(u))
}

var (
	_ service.Resource          = (*User)(nil)
	_ service.CRUDService[User] = (*Users)(nil)
)
