package challenges

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Challenges struct {
	*mongo.Collection
}

func NewChallenges(c *mongo.Collection) *Challenges {
	return &Challenges{c}
}

type Challenge struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	CreatedBy   string `json:"createdBy" bson:"createdBy"`
	CreatedDate string `json:"createdDate" bson:"createdDate"`
	StartDate   string `json:"startDate" bson:"startDate"`
	EndDate     string `json:"endDate" bson:"endDate"`
	InviteOnly  bool   `json:"inviteOnly" bson:"inviteOnly"`
}

func (c *Challenges) GetChallenges(ctx context.Context) ([]Challenge, error) {

	var challenges []Challenge

	cur, err := c.Find(ctx, bson.D{})

	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &challenges); err != nil {
		return nil, err
	}

	return challenges, err

}
