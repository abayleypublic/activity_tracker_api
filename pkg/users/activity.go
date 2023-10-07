package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/bson"
)

func (u *Users) ReadUserActivity(ctx context.Context, id service.ID, aID service.ID) (activities.Activity, error) {

	activity := activities.Activity{}
	if err := u.activities.FindResource(ctx, &activity, bson.D{{Key: "userID", Value: id}, {Key: u.activities.IDKey, Value: aID}}); err != nil {
		return activity, err
	}

	return activity, nil

}

func (u *Users) UpdateUserActivity(ctx context.Context, userID service.ID, activity activities.Activity) (activities.Activity, error) {

	if err := u.DeleteUserActivity(ctx, userID, activity.ID); err != nil {
		return activities.Activity{}, err
	}

	_, err := u.CreateUserActivity(ctx, userID, activity)
	if err != nil {
		return activities.Activity{}, err
	}

	return activity, nil

}

// CreateUserActivity adds a new activity to the user's activities.
// It returns the id of the inserted activity and an error if any occurred.
func (u *Users) CreateUserActivity(ctx context.Context, userID service.ID, activity activities.Activity) (service.ID, error) {
	return u.activities.Create(ctx, activity)
}

// DeleteUserActivity removes an activity with the given id from the user's activities.
// It returns a boolean indicating whether the deletion was successful and an error if any occurred.
func (u *Users) DeleteUserActivity(ctx context.Context, userID service.ID, activityID service.ID) error {
	return u.activities.Delete(ctx, activityID)
}
