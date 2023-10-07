package activities

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"go.mongodb.org/mongo-driver/bson"
)

func (a *Activities) ReadActivity(ctx context.Context, id service.ID) (Activity, error) {

	activity := Activity{}
	if err := a.Read(ctx, id, &activity); err != nil {
		return activity, err
	}

	return activity, nil

}

func (a *Activities) UpdateActivity(ctx context.Context, activity Activity) (Activity, error) {

	if err := a.UpdateWithCriteria(ctx, activity, bson.D{{Key: a.IDKey, Value: activity.ID}}); err != nil {
		return Activity{}, err
	}

	return activity, nil

}

// CreateUserActivity adds a new activity to the user's activities.
// It returns the id of the inserted activity and an error if any occurred.
func (a *Activities) CreateActivity(ctx context.Context, activity Activity) (service.ID, error) {
	return a.Create(ctx, activity)
}

// DeleteUserActivity removes an activity with the given id from the user's activities.
// It returns a boolean indicating whether the deletion was successful and an error if any occurred.
func (a *Activities) DeleteActivity(ctx context.Context, activityID service.ID) error {
	return a.DeleteWithCriteria(ctx, bson.D{{Key: a.IDKey, Value: activityID}})
}
