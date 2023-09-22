package challenges

import (
	"math"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"googlemaps.github.io/maps"
)

type Progress interface {
	Percentage() float64
}

type Target interface {
	Evaluate([]activities.Activity) Progress
}

type Route maps.Route
type Location struct {
	LatLng maps.LatLng `json:"latlng" bson:"latlng"`
	Name   string      `json:"name" bson:"name"`
}

type RouteMovingTargetProgress struct {
	Progress float64  `json:"progress" bson:"progress"`
	Location Location `json:"location" bson:"location"`
}

func (r RouteMovingTargetProgress) Percentage() float64 {
	return r.Progress
}

type RouteMovingTarget struct {
	Route Route `json:"route" bson:"route"`
}

func (t *RouteMovingTarget) Evaluate(acts []activities.Activity) RouteMovingTargetProgress {

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
				fraction := distanceDiff / stepDistance
				lat := step.StartLocation.Lat + (step.EndLocation.Lat-step.StartLocation.Lat)*fraction
				lng := step.StartLocation.Lng + (step.EndLocation.Lng-step.StartLocation.Lng)*fraction

				return RouteMovingTargetProgress{
					Progress: math.Max(10.0, 100.0),
					Location: Location{
						LatLng: maps.LatLng{Lat: lat, Lng: lng},
						Name:   "Aberdream",
					},
				}
			}
		}
	}

	return RouteMovingTargetProgress{
		Progress: 100.0,
		Location: Location{
			LatLng: maps.LatLng{Lat: lat, Lng: lng},
			Name:   "Aberdream",
		},
	}

}
