package targets

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
)

type TargetType string

// Progress represents the progress of a task as a percentage.
type Progress interface {
	// Percentage returns the progress as a percentage between 0 and 100.
	Percentage() float64
}

// Target represents a target to be achieved by a set of activities.
type Target interface {
	// Type returns the type of target, e.g. "routeMovingTarget"
	Type() TargetType
	// Evaluate evaluates the given activities and returns the progress towards the target.
	Evaluate(context.Context, []activities.Activity) (Progress, error)
}

type BaseTarget struct {
	TargetType TargetType `json:"type" bson:"type"`
}

func (t *BaseTarget) Type() TargetType {
	return t.TargetType
}
