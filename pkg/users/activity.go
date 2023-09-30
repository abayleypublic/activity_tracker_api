package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// TODO - implement
func (u *Users) ReadUserActivity(ctx context.Context, id uuid.ID, aID uuid.ID) (activities.Activity, error) {

	activity := activities.Activity{}
	if err := u.FindOne(ctx, bson.D{{Key: "_id", Value: id}, {Key: "activities._id", Value: aID}}).Decode(&activity); err != nil {
		return activity, err
	}

	return activity, nil

}

func (u *Users) UpdateUserActivity(ctx context.Context, userID uuid.ID, activity activities.Activity) (activities.Activity, error) {

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
func (u *Users) CreateUserActivity(ctx context.Context, userID uuid.ID, activity activities.Activity) (uuid.ID, error) {

	res, err := u.AppendAttribute(ctx, userID, "activities", activity)
	if err != nil {
		return "", err
	}

	return res, nil

}

// DeleteUserActivity removes an activity with the given id from the user's activities.
// It returns a boolean indicating whether the deletion was successful and an error if any occurred.
func (u *Users) DeleteUserActivity(ctx context.Context, userID uuid.ID, activityID uuid.ID) error {

	if err := u.RemoveAttribute(ctx, userID, activityID, "activities"); err != nil {
		return err
	}

	return nil

}
