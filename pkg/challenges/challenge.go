package challenges

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *Challenges) GetChallenge(ctx context.Context, id string) (Challenge, error) {

	var challenge Challenge

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return Challenge{}, err
	}

	if err = c.FindOne(ctx, bson.D{{Key: "_id", Value: oid}}).Decode(&challenge); err != nil {
		return Challenge{}, err
	}

	return challenge, err

}

func (c *Challenges) PostChallenge(ctx context.Context, challenge Challenge) (primitive.ObjectID, error) {

	res, err := c.InsertOne(ctx, challenge)

	if err != nil {
		return primitive.ObjectID{}, err
	}

	return res.InsertedID.(primitive.ObjectID), nil

}

func (c *Challenges) DeleteChallenge(ctx context.Context, id string) (bool, error) {

	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return false, err
	}

	res, err := c.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})

	if err != nil {
		return false, err
	}

	return res.DeletedCount == 0, nil

}
