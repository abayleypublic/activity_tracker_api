package users

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
)

// ReadUserActivities retrieves the activities of a user with the given id.
// It returns a slice of activities and an error if any occurred.
func (u *Users) ReadUserActivities(ctx context.Context, id uuid.ID) ([]activities.Activity, error) {

	user, err := u.ReadUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return user.Activities, nil

}
