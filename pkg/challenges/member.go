package challenges

import (
	"context"
	"errors"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// PutMember adds a member to a challenge. It takes a context, a challengeID and a userID as parameters.
// It returns a boolean indicating if the operation was successful and an error if any occurred.
// The function first checks if the challengeID and userID are not empty. If they are, it returns an error.
// Then it converts the challengeID and userID to a format suitable for the database.
// It then attempts to update the challenge document in the database by adding the userID to the members field.
// If the document is not found, it returns an error. If any other error occurs during the update, it is returned.
func (c *Challenges) PutMember(ctx context.Context, challengeID uuid.ID, userID uuid.ID) (bool, error) {
	if challengeID == "" || userID == "" {
		return false, errors.New("challengeID and userID cannot be empty")
	}

	cid, err := uuid.ConvertID(challengeID)
	if err != nil {
		return false, err
	}

	uid, err := uuid.ConvertID(userID)
	if err != nil {
		return false, err
	}

	result, err := c.UpdateOne(ctx, bson.D{{Key: "_id", Value: cid}}, bson.D{{Key: "$push", Value: bson.D{{Key: "members", Value: uid}}}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("document not found")
		}
		return false, err
	}

	return result.ModifiedCount == 1, nil
}

// DeleteMember removes a member from a challenge. It takes a context, a challengeID and a userID as parameters.
// It returns a boolean indicating if the operation was successful and an error if any occurred.
// The function first checks if the challengeID and userID are not empty. If they are, it returns an error.
// Then it converts the challengeID and userID to a format suitable for the database.
// It then attempts to update the challenge document in the database by removing the userID from the members field.
// If the document is not found, it returns an error. If any other error occurs during the update, it is returned.
func (c *Challenges) DeleteMember(ctx context.Context, challengeID uuid.ID, userID uuid.ID) (bool, error) {
	if challengeID == "" || userID == "" {
		return false, errors.New("challengeID and userID cannot be empty")
	}

	cid, err := uuid.ConvertID(challengeID)
	if err != nil {
		return false, err
	}

	uid, err := uuid.ConvertID(userID)
	if err != nil {
		return false, err
	}

	result, err := c.UpdateOne(ctx, bson.D{{Key: "_id", Value: cid}}, bson.D{{Key: "$pull", Value: bson.D{{Key: "members", Value: uid}}}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("document not found")
		}
		return false, err
	}

	return result.ModifiedCount == 1, nil
}
