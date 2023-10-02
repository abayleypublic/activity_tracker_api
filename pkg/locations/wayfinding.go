package locations

import (
	"context"

	"github.com/uber/h3-go/v4"
	"googlemaps.github.io/maps"
)

type LatLng h3.LatLng

type Waypoint struct {
	LatLng LatLng `json:"latlng" bson:"latlng"`
}

// Returns distance between 2 waypoints in km
func (w Waypoint) DistanceTo(other Waypoint) float64 {
	return h3.GreatCircleDistanceKm(h3.LatLng(w.LatLng), h3.LatLng(other.LatLng))
}

type Waypoints []Waypoint

func (w Waypoints) Last() Waypoint {
	return w[len(w)-1]
}

type Location struct {
	LatLng LatLng `json:"latlng" bson:"latlng"`
	Name   string `json:"name" bson:"name"`
}

func LocationFromLatLng(latlng LatLng) (Location, error) {

	name, err := GetLocationName(context.Background(), maps.LatLng(latlng))
	if err != nil {
		return Location{}, err
	}

	return Location{
		LatLng: latlng,
		Name:   name,
	}, nil
}

// TODO Implement functions from https://www.movable-type.co.uk/scripts/latlong.html to get more granular location data

// Iterates over the waypoints and returns the location when the distance (total distance travelled by user) is reached
func (w Waypoints) GetLocation(distance float64) (Location, error) {

	var distanceSum float64
	var previous *Waypoint
	// Iterate over each leg of the route and calculate the distance sum
	for _, waypoint := range w {
		if previous == nil {
			previous = &waypoint
			continue
		}

		distanceSum += previous.DistanceTo(waypoint)

		// Check if the total distance has been reached
		if distanceSum >= distance {
			return LocationFromLatLng(previous.LatLng)
		}
	}

	return LocationFromLatLng(w.Last().LatLng)
}
