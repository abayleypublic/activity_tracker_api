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
	ID          string
	Name        string
	Description string
	CreatedBy   string
	StartDate   string
	EndDate     string
	InviteOnly  bool
}
