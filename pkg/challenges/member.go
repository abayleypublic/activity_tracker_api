package challenges

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *Challenges) AddMember(ctx context.Context, challengeID service.ID, userID service.ID) (service.ID, error) {

	res, err := c.AppendAttribute(ctx, challengeID, "members", userID)
	if err != nil {
		return "", err
	}

	return res, err
}

func (c *Challenges) DeleteMember(ctx context.Context, challengeID service.ID, userID service.ID) error {

	result, err := c.UpdateOne(ctx, bson.D{{Key: "_id", Value: challengeID}}, bson.D{{Key: "$pull", Value: bson.D{{Key: "members", Value: userID}}}})
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
