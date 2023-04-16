package challenges

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Challenges struct {
	*mongo.Collection
}

func NewChallenges(c *mongo.Collection) *Challenges {
	return &Challenges{c}
}

type Challenge struct {
	id          string
	name        string
	description string
	createdBy   string
	startDate   string
	endDate     string
	inviteOnly  bool
}
