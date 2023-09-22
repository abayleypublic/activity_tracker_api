package challenges

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Challenges wraps a MongoDB collection of challenges.
type Challenges struct {
	*mongo.Collection
}

// NewChallenges creates a new Challenges instance with the provided MongoDB collection.
func NewChallenges(c *mongo.Collection) *Challenges {
	return &Challenges{c}
}

// PartialChallenge represents a challenge with a subset of its fields.
type PartialChallenge struct {
	ID          uuid.ID `json:"id" bson:"_id,omitempty"`
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	CreatedBy   string  `json:"createdBy" bson:"createdBy"`
	CreatedDate string  `json:"createdDate" bson:"createdDate"`
	StartDate   string  `json:"startDate" bson:"startDate"`
	EndDate     string  `json:"endDate" bson:"endDate"`
	Public      bool    `json:"public" bson:"public"`
	InviteOnly  bool    `json:"inviteOnly" bson:"inviteOnly"`
}

// Challenge represents a full challenge, including its members.
type Challenge struct {
	PartialChallenge
	Members []Member `json:"members" bson:"members"`
}

// GetChallenges retrieves all challenges from the MongoDB collection.
// It returns a slice of PartialChallenge and an error if there is any.
func (c *Challenges) GetChallenges(ctx context.Context) ([]PartialChallenge, error) {

	var challenges []PartialChallenge

	cur, err := c.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &challenges); err != nil {
		return nil, err
	}

	return challenges, err

}
