package challenges

import (
	"context"
	"errors"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddMember adds a member to a challenge. It takes a context, a challengeID and a userID as parameters.
// It returns a boolean indicating if the operation was successful and an error if any occurred.
// The function first checks if the challengeID and userID are not empty. If they are, it returns an error.
// Then it converts the challengeID and userID to a format suitable for the database.
// It then attempts to update the challenge document in the database by adding the userID to the members field.
// If the document is not found, it returns an error. If any other error occurs during the update, it is returned.
func (c *Challenges) AddMember(ctx context.Context, challengeID uuid.ID, userID uuid.ID) error {
	if challengeID == "" || userID == "" {
		return errors.New("challengeID and userID cannot be empty")
	}

	result, err := c.UpdateOne(ctx, bson.D{{Key: "_id", Value: challengeID}}, bson.D{{Key: "$push", Value: bson.D{{Key: "members", Value: userID}}}})
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return ErrResourceNotFound
		}
	}

	if result.ModifiedCount != 1 {
		return ErrResourceNotFound
	}

	return err
}

// DeleteMember removes a member from a challenge. It takes a context, a challengeID and a userID as parameters.
// It returns a boolean indicating if the operation was successful and an error if any occurred.
// The function first checks if the challengeID and userID are not empty. If they are, it returns an error.
// Then it converts the challengeID and userID to a format suitable for the database.
// It then attempts to update the challenge document in the database by removing the userID from the members field.
// If the document is not found, it returns an error. If any other error occurs during the update, it is returned.
func (c *Challenges) DeleteMember(ctx context.Context, challengeID uuid.ID, userID uuid.ID) error {
	if challengeID == "" || userID == "" {
		return errors.New("challengeID and userID cannot be empty")
	}

	result, err := c.UpdateOne(ctx, bson.D{{Key: "_id", Value: challengeID}}, bson.D{{Key: "$pull", Value: bson.D{{Key: "members", Value: userID}}}})
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return ErrResourceNotFound
		}
	}

	if result.ModifiedCount != 1 {
		return ErrResourceNotFound
	}

	return err
}
