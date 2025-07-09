package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	challengesKey = "challenges"
)

// Returns either a User or PartialUser depending on whether the user is the actor.
func (u *Users) ReadUser(ctx context.Context, id service.ID) (PartialUser, error) {

	user := PartialUser{}
	if err := u.ReadRaw(ctx, service.ID(id), &user); err != nil {
		return user, err
	}
	return user, nil

}

func (u *Users) JoinChallenge(ctx context.Context, userID service.ID, challengeID service.ID) (service.ID, error) {

	res, err := u.AppendAttribute(ctx, userID, challengesKey, challengeID)
	if err != nil {
		return "", err
	}

	return res, err
}

func (u *Users) LeaveChallenge(ctx context.Context, userID service.ID, challengeID service.ID) error {

	result, err := u.UpdateOne(ctx, bson.D{{Key: "_id", Value: userID}}, bson.D{{Key: "$pull", Value: bson.D{{Key: challengesKey, Value: challengeID}}}})
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return service.ErrResourceNotFound
		}

		return service.ErrUnknownError
	}

	if result.ModifiedCount != 1 {
		return service.ErrResourceNotFound
	}

	return err
}

func (u *Users) IsUserMember(ctx context.Context, userID service.ID, challengeID service.ID) (bool, error) {

	user := User{}
	if err := u.Read(ctx, userID, &user); err != nil {
		return false, err
	}

	for _, id := range user.Challenges {
		if id == challengeID {
			return true, nil
		}
	}

	return false, nil

}
