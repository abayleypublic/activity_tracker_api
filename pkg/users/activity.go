package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostUserActivity adds a new activity to the user's activities.
// It returns the id of the inserted activity and an error if any occurred.
func (u *Users) PostUserActivity(ctx context.Context, activity activities.Activity) (uuid.ID, error) {

	res, err := u.InsertOne(ctx, activity)

	if err != nil {
		return "", err
	}

	return uuid.ID(res.InsertedID.(primitive.ObjectID).String()), nil

}

// DeleteUserActivity removes an activity with the given id from the user's activities.
// It returns a boolean indicating whether the deletion was successful and an error if any occurred.
func (u *Users) DeleteUserActivity(ctx context.Context, id uuid.ID) (bool, error) {

	oid, err := uuid.ConvertID(id)

	if err != nil {
		return false, err
	}

	res, err := u.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})

	if err != nil {
		return false, err
	}

	return res.DeletedCount == 1, nil

}
