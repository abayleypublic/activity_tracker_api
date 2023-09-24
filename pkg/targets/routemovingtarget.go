package targets

import (
	"context"
	"math"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/locations"
	"googlemaps.github.io/maps"
)

var (
	_ Target   = (*RouteMovingTarget)(nil)
	_ Progress = (*RouteMovingTargetProgress)(nil)
)

const (
	RouteMovingTargetType TargetType = "routeMovingTarget"
)

type Route maps.Route

type RouteMovingTargetProgress struct {
	Progress float64            `json:"progress" bson:"progress"`
	Location locations.Location `json:"location" bson:"location"`
}

func (r RouteMovingTargetProgress) Percentage() float64 {
	return r.Progress
}

type RouteMovingTarget struct {
	BaseTarget
	Route Route `json:"route" bson:"route"`
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

	var distanceSum float64
	// Iterate over each leg of the route and calculate the distance sum
	for _, leg := range t.Route.Legs {
		for _, step := range leg.Steps {
			stepDistance := float64(step.Distance.Meters) / 1000
			distanceSum += stepDistance

			// Check if the total distance has been reached
			if distanceSum >= distance {
				distanceDiff := distanceSum - distance
				fraction := (distanceDiff / stepDistance) * 100

				latlng := locations.LatLng(step.StartLocation)
				loc, err := locations.LocationFromLatLng(ctx, latlng)
				if err != nil {
					return nil, err
				}

				return RouteMovingTargetProgress{
					Progress: math.Max(fraction, 100.0),
					Location: loc,
				}, nil
			}
		}
	}

	lastLeg := t.Route.Legs[len(t.Route.Legs)-1]
	latlng := locations.LatLng(lastLeg.EndLocation)
	loc, err := locations.LocationFromLatLng(ctx, latlng)
	if err != nil {
		return nil, err
	}

	return RouteMovingTargetProgress{
		Progress: 100.0,
		Location: loc,
	}, nil

}
