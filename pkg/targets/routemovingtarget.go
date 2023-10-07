package targets

import (
	"context"
	"errors"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets/locations"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	_ Target   = (*RouteMovingTarget)(nil)
	_ Progress = (*RouteMovingTargetProgress)(nil)
)

const (
	RouteMovingTargetType TargetType = "routeMovingTarget"
)

var (
	ErrFindingLocation = errors.New("error finding location")
)

// Can't use Route as Google don't allow caching / storage of data from the Directions API
type Route struct {
	locations.Waypoints `json:"waypoints" bson:"waypoints"`
}

func (r *Route) MarshalBSON() ([]byte, error) {
	type RawRoute Route
	if r.Waypoints == nil {
		r.Waypoints = make(locations.Waypoints, 0)
	}

	return bson.Marshal((*RawRoute)(r))
}

type RouteMovingTargetProgress struct {
	Percent  float64            `json:"percent" bson:"percent"`
	Location locations.Location `json:"location" bson:"location"`
}

func (r RouteMovingTargetProgress) Percentage() float64 {
	return r.Percent
}

type RouteMovingTarget struct {
	BaseTarget `bson:",inline"`
	Route      Route `json:"route" bson:"route"`
}

func (t *RouteMovingTarget) Type() TargetType {
	return RouteMovingTargetType
}

func (t *RouteMovingTarget) Evaluate(ctx context.Context, acts []activities.Activity) (Progress, error) {

	// Distance is the distance travelled by the user
	var distance float64 = 0
	for _, act := range acts {
		if _, ok := activities.Moving[act.Type]; ok {
			distance += act.Value
		}
	}

	loc, err := t.Route.GetLocation(distance)
	if err != nil {
		return nil, ErrFindingLocation
	}

	return RouteMovingTargetProgress{
		Percent:  100.0,
		Location: loc,
	}, nil

}
