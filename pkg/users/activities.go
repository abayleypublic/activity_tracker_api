package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
)

// ReadUserActivities retrieves the activities of a user with the given id.
// It returns a slice of activities and an error if any occurred.
func (u *Users) ReadUserActivities(ctx context.Context, id service.ID) ([]activities.Activity, error) {

	activities := []activities.Activity{}
	err := u.ReadAttribute(ctx, id, "activities", &activities)
	if err != nil {
		return nil, err
	}

	return activities, nil

}
