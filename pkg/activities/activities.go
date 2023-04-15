package activities

import (
	"github.com/monzo/typhon"
	"go.mongodb.org/mongo-driver/mongo"
)

type Activities struct {
	*mongo.Collection
}

func NewActivities(c *mongo.Collection) *Activities {
	return &Activities{c}
}

type Activity struct {
	string
}

func (a *Activities) GetActivities(req typhon.Request) typhon.Response {

	return req.Response("OK")

}
