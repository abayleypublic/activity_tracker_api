package challenges

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrInvalidTarget     = errors.New("invalid target type")
	ErrParseTarget       = errors.New("error parsing target")
	ErrChallengeNotFound = errors.New("challenge not found")
)

// Challenges wraps a MongoDB collection of challenges.
type Challenges struct {
	*mongo.Collection
}

// NewChallenges creates a new Challenges instance with the provided MongoDB collection.
func NewChallenges(c *mongo.Collection) *Challenges {
	return &Challenges{c}
}

// PartialChallenge represents a challenge with a subset of its fields mainly intended for parsing a request from the UI
type BaseChallenge struct {
	ID          uuid.ID `json:"id" bson:"_id,omitempty"`
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	StartDate   string  `json:"startDate" bson:"startDate"`
	EndDate     string  `json:"endDate" bson:"endDate"`
	Public      bool    `json:"public" bson:"public"`
	InviteOnly  bool    `json:"inviteOnly" bson:"inviteOnly"`
}

// PartialChallenge builds on BaseChallenge by adding the creator and target fields.
type PartialChallenge struct {
	BaseChallenge `bson:",inline"`
	CreatedBy     uuid.ID `json:"createdBy" bson:"createdBy"`
	CreatedDate   string  `json:"createdDate" bson:"createdDate"`
}

type PartialChallengeWithTarget struct {
	PartialChallenge `bson:",inline"`
	Target           targets.Target `json:"target" bson:"target"`
}

func (c *PartialChallengeWithTarget) UnmarshalJSON(b []byte) error {

	type RawChallenge struct {
		PartialChallenge
		Target targets.BaseTarget `json:"target"`
	}

	// Parse rawMessage
	rawChallenge := &RawChallenge{}
	if err := json.Unmarshal(b, rawChallenge); err != nil {
		return err
	}
	c.PartialChallenge = rawChallenge.PartialChallenge

	switch rawChallenge.Target.Type() {
	case targets.RouteMovingTargetType:
		type RouteMovingTargetChallenge struct {
			Target targets.RouteMovingTarget `json:"target"`
		}
		targetChallenge := &RouteMovingTargetChallenge{}
		if err := json.Unmarshal(b, targetChallenge); err != nil {
			return err
		}
		c.Target = &targetChallenge.Target
	default:
		return ErrInvalidTarget
	}
	return nil
}

// Challenge represents a full challenge, including its members.
type Challenge struct {
	PartialChallengeWithTarget `bson:",inline"`
	Members                    []Member `json:"members" bson:"members"`
}

// ReadChallenges retrieves all challenges from the MongoDB collection.
// It returns a slice of PartialChallenge and an error if there is any.
func (c *Challenges) ReadChallenges(ctx context.Context) ([]PartialChallenge, error) {

	cur, err := c.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var challenges []PartialChallenge
	if err = cur.All(ctx, &challenges); err != nil {
		return challenges, err
	}

	return challenges, err

}
