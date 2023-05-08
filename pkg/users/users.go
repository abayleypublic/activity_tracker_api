package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Users struct {
	*mongo.Collection
}

func NewUsers(c *mongo.Collection) *Users {
	return &Users{c}
}

type User struct {
	ID          string                `json:"id" bson:"_id"`
	CreatedDate string                `json:"createdDate" bson:"createdDate"`
	FirstName   string                `json:"firstName,omitempty" bson:"firstName"`
	LastName    string                `json:"lastName,omitempty" bson:"lastName"`
	Email       string                `json:"email,omitempty" bson:"email"`
	Bio         string                `json:"bio,omitempty" bson:"bio"`
	Challenges  []string              `json:"challenges,omitempty" bson:"challenges"`
	Activities  []activities.Activity `json:"activities,omitempty" bson:"activities"`
}

type PartialUser struct {
	ID        string `json:"id" bson:"_id"`
	FirstName string `json:"firstName,omitempty" bson:"firstName"`
	LastName  string `json:"lastName,omitempty" bson:"lastName"`
}

func (u *Users) GetUsers(ctx context.Context) ([]User, error) {

	cur, err := u.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	users := []User{}
	if err = cur.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil

}
