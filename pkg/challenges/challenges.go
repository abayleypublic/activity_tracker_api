package challenges

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
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
	*service.Service[Challenge]
}

// NewChallenges creates a new Challenges instance with the provided MongoDB collection.
func NewChallenges(c *mongo.Collection) *Challenges {
	return &Challenges{c, service.New[Challenge](c)}
}

// PartialChallenge represents a challenge with a subset of its fields mainly intended for parsing a request from the UI
type BaseChallenge struct {
	ID          uuid.ID   `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	StartDate   time.Time `json:"startDate" bson:"startDate"`
	EndDate     time.Time `json:"endDate" bson:"endDate"`
	Public      bool      `json:"public" bson:"public"`
	InviteOnly  bool      `json:"inviteOnly" bson:"inviteOnly"`
}

func (bc BaseChallenge) GetID() uuid.ID {
	return bc.ID
}

// PartialChallenge builds on BaseChallenge by adding the creator and target fields.
type PartialChallenge struct {
	BaseChallenge `bson:",inline"`
	CreatedBy     uuid.ID   `json:"createdBy" bson:"createdBy"`
	CreatedDate   time.Time `json:"createdDate" bson:"createdDate"`
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
	Members                    []Member `json:"members,omitempty" bson:"members"`
}

func (c *Challenge) MarshalBSON() ([]byte, error) {
	type RawChallenge Challenge
	if c.Members == nil {
		c.Members = make([]Member, 0)
	}

	return bson.Marshal((*RawChallenge)(c))
}

var (
	_ service.Resource               = (*Challenge)(nil)
	_ service.CRUDService[Challenge] = (*Challenges)(nil)
)
