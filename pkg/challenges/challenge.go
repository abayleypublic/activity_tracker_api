package challenges

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetChallenge retrieves a challenge from the database using the provided ID.
// It returns the retrieved challenge and any error encountered during the operation.
func (c *Challenges) GetChallenge(ctx context.Context, id uuid.ID) (Challenge, error) {

	var challenge Challenge

	oid, err := uuid.ConvertID(id)

	if err != nil {
		return Challenge{}, err
	}

	if err = c.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&challenge); err != nil {
		return Challenge{}, err
	}

	return challenge, err

}

// PostChallenge adds a new challenge to the database.
// It returns the ID of the newly inserted challenge and any error encountered during the operation.
func (c *Challenges) PostChallenge(ctx context.Context, challenge Challenge) (uuid.ID, error) {

	res, err := c.InsertOne(ctx, challenge)

	if err != nil {
		return "", err
	}

	return uuid.ID(res.InsertedID.(primitive.ObjectID).String()), nil

}

// DeleteChallenge removes a challenge from the database using the provided ID.
// It returns a boolean indicating whether the deletion was successful and any error encountered during the operation.
func (c *Challenges) DeleteChallenge(ctx context.Context, id uuid.ID) (bool, error) {

	oid, err := uuid.ConvertID(id)
	if err != nil {
		return false, err
	}

	res, err := c.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})

	if err != nil {
		return false, err
	}

	return res.DeletedCount == 1, nil

}
