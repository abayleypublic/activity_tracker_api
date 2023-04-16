package users

import (
	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"go.mongodb.org/mongo-driver/mongo"
)

type Users struct {
	*mongo.Collection
}

func NewUsers(c *mongo.Collection) *Users {
	return &Users{c}
}

type User struct {
	id         string
	firstName  string
	lastName   string
	email      string
	bio        string
	challenges []string
	activities []activities.Activity
}
