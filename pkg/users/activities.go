package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// GetUserActivities retrieves the activities of a user with the given id.
// It returns a slice of activities and an error if any occurred.
func (u *Users) GetUserActivities(ctx context.Context, id uuid.ID) ([]activities.Activity, error) {

	var user User
	if err := u.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&user); err != nil {
		return nil, err
	}

	return user.Activities, nil

}
