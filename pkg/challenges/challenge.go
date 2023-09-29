package challenges

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReadChallenge retrieves a challenge from the database using the provided ID.
// It returns the retrieved challenge and any error encountered during the operation.
func (c *Challenges) ReadChallenge(ctx context.Context, id uuid.ID) (Challenge, error) {
	challenge := Challenge{}
	if err := c.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&challenge); err != nil {
		return challenge, err
	}

	return challenge, nil

}

// CreateChallenge adds a new challenge to the database.
// It returns the ID of the newly inserted challenge and any error encountered during the operation.
func (c *Challenges) CreateChallenge(ctx context.Context, challenge Challenge) (uuid.ID, error) {
	res, err := c.InsertOne(ctx, challenge)
	if err != nil {
		return "", err
	}

	return uuid.ID(res.InsertedID.(primitive.ObjectID).String()), nil

}

func (c *Challenges) UpdateChallenge(ctx context.Context, challenge Challenge) error {
	opts := options.Update().SetUpsert(true)
	res, err := c.UpdateOne(ctx, bson.D{{Key: "$set", Value: challenge}}, opts)
	if res.UpsertedCount != 1 {
		return ErrChallengeNotFound
	}

	return err
}

// DeleteChallenge removes a challenge from the database using the provided ID.
// It returns a boolean indicating whether the deletion was successful and any error encountered during the operation.
func (c *Challenges) DeleteChallenge(ctx context.Context, id uuid.ID) error {

	res, err := c.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if res.DeletedCount != 1 {
		return ErrChallengeNotFound
	}

	return err

}
