package targets

import (
	"context"
	"errors"
	"math"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/locations"
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
	Percent         float64            `json:"percent" bson:"percent"`
	DistanceCovered float64            `json:"distanceCovered" bson:"distanceCovered"`
	Location        locations.Location `json:"location" bson:"location"`
}

func (r RouteMovingTargetProgress) Percentage() float64 {
	return r.Percent
}

type RouteMovingTarget struct {
	BaseTarget    `bson:",inline"`
	Route         Route   `json:"route" bson:"route"`
	TotalDistance float64 `json:"totalDistance" bson:"totalDistance"`
}

func (t *RouteMovingTarget) MarshalBSON() ([]byte, error) {

	type RawRouteMovingTarget RouteMovingTarget

	if len(t.Route.Waypoints) < 2 {
		return bson.Marshal((*RawRouteMovingTarget)(t))
	}

	previous := t.Route.Waypoints[0]

	var distanceSum float64 = 0
	for _, waypoint := range t.Route.Waypoints[1:] {
		distanceSum += previous.DistanceTo(waypoint)
		previous = waypoint
	}
	t.TotalDistance = distanceSum

	return bson.Marshal((*RawRouteMovingTarget)(t))
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

	var percent float64 = 0
	if distance > 0 && t.TotalDistance > 0 {
		percent = math.Min((distance/t.TotalDistance)*100, 100)
	}

	return RouteMovingTargetProgress{
		Percent:         percent,
		DistanceCovered: distance,
		Location:        loc,
	}, nil

}
