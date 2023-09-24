package users

import (
	"context"
	"errors"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO - implement
func (u *Users) ReadUserActivity(ctx context.Context, id uuid.ID, aID uuid.ID) (activities.Activity, error) {

	activity := activities.Activity{}
	if err := u.FindOne(ctx, bson.D{{Key: "_id", Value: id}, {Key: "activities._id", Value: aID}}).Decode(&activity); err != nil {
		return activity, err
	}

	return activity, nil

}

// TODO - implement
func (u *Users) UpdateUserActivity(ctx context.Context, userID uuid.ID, activity activities.Activity) (activities.Activity, error) {

	if _, err := u.DeleteUserActivity(ctx, userID, activity.ID); err != nil {
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

	uid, err := uuid.ConvertID(userID)
	if err != nil {
		return "", err
	}

	result, err := u.UpdateOne(ctx, bson.D{{Key: "_id", Value: uid}}, bson.D{{Key: "$push", Value: bson.D{{Key: "activities", Value: activity}}}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New("user not found")
		}
		return "", err
	}

	return uuid.ID(result.UpsertedID.(primitive.ObjectID).String()), nil

}

// DeleteUserActivity removes an activity with the given id from the user's activities.
// It returns a boolean indicating whether the deletion was successful and an error if any occurred.
func (u *Users) DeleteUserActivity(ctx context.Context, userID uuid.ID, activityID uuid.ID) (bool, error) {

	uid, err := uuid.ConvertID(userID)
	if err != nil {
		return false, err
	}

	aid, err := uuid.ConvertID(activityID)
	if err != nil {
		return false, err
	}

	result, err := u.UpdateOne(ctx, bson.D{{Key: "_id", Value: uid}}, bson.D{{Key: "$pull", Value: bson.D{{Key: "activities._id", Value: aid}}}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("user not found")
		}
		return false, err
	}

	return result.ModifiedCount == 1, nil

}
