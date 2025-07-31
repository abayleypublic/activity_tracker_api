package targets

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrInvalidTarget = errors.New("invalid target")
	ErrSyntaxError   = errors.New("syntax error")
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

type RawTarget struct {
	BaseTarget `bson:",inline"`
	RealTarget Target `json:"-" bson:"-"`
}

func (t *RawTarget) UnmarshalBSON(b []byte) error {
	// Get the type of target
	raw := bson.Raw(b)
	if err := raw.Lookup("type").Unmarshal(&t.TargetType); err != nil {
		return ErrInvalidTarget
	}

	// Get a pointer to the target type
	rt := resolveType(t.TargetType)
	if rt == nil {
		// If pointer is nil, an invalid target type was supplied
		return ErrInvalidTarget
	}

	// Get the type of the target type
	tar := reflect.TypeOf(rt)
	// Make a new pointer to the target type
	v := reflect.New(tar.Elem())
	// Unmarshal the bson into the new pointer
	ptr := v.Interface()
	if err := bson.Unmarshal(b, ptr); err != nil {
		return ErrSyntaxError
	}
	t.RealTarget = ptr.(Target)

	return nil
}

func (t *RawTarget) UnmarshalJSON(b []byte) error {
	raw := map[string]interface{}{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	targetType, ok := raw["type"]
	if !ok {
		return ErrInvalidTarget
	}
	t.TargetType = TargetType(targetType.(string))

	// Get a pointer to the target type
	rt := resolveType(t.TargetType)
	if rt == nil {
		// If pointer is nil, an invalid target type was supplied
		return ErrInvalidTarget
	}

	// Get the type of the target type
	tar := reflect.TypeOf(rt)
	// Make a new pointer to the target type
	v := reflect.New(tar.Elem())
	// Unmarshal the bson into the new pointer
	ptr := v.Interface()
	if err := json.Unmarshal(b, ptr); err != nil {
		return ErrSyntaxError
	}
	t.RealTarget = ptr.(Target)

	return nil
}

// resolveType returns a new instance of the target type based on the provided TargetType.
func resolveType(targetType TargetType) Target {
	var target Target
	switch targetType {
	case RouteMovingTargetType:
		target = &RouteMovingTarget{}
	}

	return target
}
