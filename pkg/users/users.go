// Package users provides the structure and methods for handling user data in the application.
package users

import (
	"errors"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrResourceNotFound  = errors.New("resource not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// Users is a wrapper around a MongoDB collection of users.
type Users struct {
	*mongo.Collection
	*service.Service[User]
}

// NewUsers creates a new Users instance with the provided MongoDB collection.
func NewUsers(c *mongo.Collection) *Users {
	return &Users{c, service.New[User](c)}
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
	CreatedDate *service.Time         `json:"createdDate" bson:"createdDate"`
	Email       string                `json:"email,omitempty" bson:"email"`
	Bio         string                `json:"bio,omitempty" bson:"bio"`
	Activities  []activities.Activity `json:"activities,omitempty" bson:"activities"`
}

func (u User) GetCreatedDate() service.Time {
	return *u.CreatedDate
}

func (u *User) MarshalBSON() ([]byte, error) {
	type RawUser User
	if u.Activities == nil {
		u.Activities = make([]activities.Activity, 0)
	}

	return bson.Marshal((*RawUser)(u))
}

var (
	_ service.Resource          = (*User)(nil)
	_ service.CRUDService[User] = (*Users)(nil)
)
