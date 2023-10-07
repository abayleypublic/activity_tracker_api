package members

import (
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
)

// Member is a type that embeds the User type from the users package.
// It represents a user who is a member of a challenge.
type Member struct {
	users.PartialUser `bson:",inline"`
	CreatedDate       *service.Time    `json:"createdDate" bson:"createdDate"`
	Progress          targets.Progress `json:"progress" bson:"progress"`
}
