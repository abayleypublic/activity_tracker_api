package challenges

import (
	"encoding/json"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Challenges wraps a MongoDB collection of challenges.
type Challenges struct {
	*mongo.Collection
	*service.MongoDBService[Challenge]
	users *users.Users
}

// NewChallenges creates a new Challenges instance with the provided MongoDB collection.
func NewChallenges(c *mongo.Collection, u *users.Users) *Challenges {
	return &Challenges{c, service.New[Challenge](c), u}
}

// PartialChallenge represents a challenge with a subset of its fields mainly intended for parsing a request from the UI
type BaseChallenge struct {
	ID          service.ID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string     `json:"name" bson:"name"`
	Description string     `json:"description" bson:"description"`
	StartDate   time.Time  `json:"startDate" bson:"startDate"`
	EndDate     time.Time  `json:"endDate" bson:"endDate"`
	Public      bool       `json:"public" bson:"public"`
	InviteOnly  bool       `json:"inviteOnly" bson:"inviteOnly"`
}

func (bc BaseChallenge) GetID() service.ID {
	return bc.ID
}

// PartialChallenge builds on BaseChallenge by adding the creator and target fields.
type PartialChallenge struct {
	BaseChallenge `bson:",inline"`
	CreatedBy     service.ID    `json:"createdBy" bson:"createdBy"`
	CreatedDate   *service.Time `json:"createdDate" bson:"createdDate"`
}

type PartialChallengeWithTarget struct {
	PartialChallenge `bson:",inline"`
	Target           targets.Target `json:"target" bson:"target"`
}

type RawChallenge struct {
	PartialChallenge `bson:",inline"`
	Target           targets.RawTarget `json:"target" bson:"target"`
}

func (c *PartialChallengeWithTarget) UnmarshalJSON(b []byte) error {
	rawChallenge := &RawChallenge{}
	if err := json.Unmarshal(b, rawChallenge); err != nil {
		return err
	}
	c.PartialChallenge = rawChallenge.PartialChallenge
	c.Target = rawChallenge.Target.RealTarget
	return nil
}

func (c *PartialChallengeWithTarget) UnmarshalBSON(b []byte) error {
	rawChallenge := &RawChallenge{}
	if err := bson.Unmarshal(b, rawChallenge); err != nil {
		return err
	}
	c.PartialChallenge = rawChallenge.PartialChallenge
	c.Target = rawChallenge.Target.RealTarget
	return nil
}

// Challenge represents a full challenge, including its members.
type Challenge struct {
	PartialChallengeWithTarget `bson:",inline"`
	CreatedDate                *service.Time `json:"createdDate" bson:"createdDate"`
}

func (c Challenge) GetCreatedDate() service.Time {
	return *c.CreatedDate
}

func (c Challenge) CanBeReadBy(userID service.ID, admin bool) bool {
	return true
}

func (c Challenge) CanBeUpdatedBy(userID service.ID, admin bool) bool {
	return c.CreatedBy == userID || admin
}

func (c Challenge) CanBeDeletedBy(userID service.ID, admin bool) bool {
	return c.CreatedBy == userID || admin
}

var (
	_ service.Resource               = (*Challenge)(nil)
	_ service.CRUDService[Challenge] = (*Challenges)(nil)
)
