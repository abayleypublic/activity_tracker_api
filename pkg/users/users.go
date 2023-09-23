// Package users provides the structure and methods for handling user data in the application.
package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Users is a wrapper around a MongoDB collection of users.
type Users struct {
	*mongo.Collection
}

// NewUsers creates a new Users instance with the provided MongoDB collection.
func NewUsers(c *mongo.Collection) *Users {
	return &Users{c}
}

// PartialUser represents a user with only the ID, first name, and last name fields.
type PartialUser struct {
	ID        uuid.ID `json:"id" bson:"_id"`
	FirstName string  `json:"firstName,omitempty" bson:"firstName"`
	LastName  string  `json:"lastName,omitempty" bson:"lastName"`
}

// User represents a full user with all fields, including activities.
type User struct {
	PartialUser
	CreatedDate string                `json:"createdDate" bson:"createdDate"`
	Email       string                `json:"email,omitempty" bson:"email"`
	Bio         string                `json:"bio,omitempty" bson:"bio"`
	Activities  []activities.Activity `json:"activities,omitempty" bson:"activities"`
}

// ReadUsers retrieves all users from the MongoDB collection.
// It returns a slice of PartialUser instances and any error encountered.
func (u *Users) ReadUsers(ctx context.Context) ([]PartialUser, error) {

	cur, err := u.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	users := []PartialUser{}
	if err = cur.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil

}
