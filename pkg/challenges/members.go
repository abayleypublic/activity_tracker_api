package challenges

import (
	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
)

type Member struct {
	users.PartialUser
	progress activities.Activity
}
