package locations

import (
	"context"
	"math"

	"github.com/uber/h3-go/v4"
	"googlemaps.github.io/maps"
)

type LatLng struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
}

func (l LatLng) AsRadians() LatLng {
	return LatLng{
		Lat: l.Lat * math.Pi / 180,
		Lng: l.Lng * math.Pi / 180,
	}
}

type Waypoint struct {
	LatLng LatLng `json:"latlng" bson:"latlng"`
}

// Returns distance between 2 waypoints in km
func (w Waypoint) DistanceTo(other Waypoint) float64 {
	return h3.GreatCircleDistanceKm(h3.LatLng(w.LatLng), h3.LatLng(other.LatLng))
}

type Waypoints []Waypoint

func (w Waypoints) First() Waypoint {
	return w[0]
}

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

func getNewCoordinates(start LatLng, end LatLng, distance float64) LatLng {
	const earthRadius = 6371
	startRad := start.AsRadians()
	endRad := end.AsRadians()

	// Calculate bearing
	deltaLng := endRad.Lng - startRad.Lng
	y := math.Sin(deltaLng) * math.Cos(endRad.Lat)
	x := math.Cos(startRad.Lat)*math.Sin(endRad.Lat) - math.Sin(startRad.Lat)*math.Cos(endRad.Lat)*math.Cos(deltaLng)
	bearing := math.Atan2(y, x)

	// Calculate new latitude
	newLat := math.Asin(math.Sin(startRad.Lat)*math.Cos(distance/earthRadius) +
		math.Cos(startRad.Lat)*math.Sin(distance/earthRadius)*math.Cos(bearing))

	// Calculate new longitude
	newLon := startRad.Lng + math.Atan2(math.Sin(bearing)*math.Sin(distance/earthRadius)*math.Cos(startRad.Lat),
		math.Cos(distance/earthRadius)-math.Sin(startRad.Lat)*math.Sin(newLat))

	// Return latlng as degrees
	return LatLng{
		Lat: newLat * 180 / math.Pi,
		Lng: newLon * 180 / math.Pi,
	}
}

// Iterates over the waypoints and returns the location when the distance (total distance travelled by user) is reached
func (w Waypoints) GetLocation(distance float64) (Location, error) {
	previous := w[0]

	var distanceSum float64 = 0
	// Iterate over each leg of the route and calculate the distance sum
	for _, waypoint := range w[1:] {

		diff := previous.DistanceTo(waypoint)
		distanceSum += diff

		// Check if the total distance has been reached
		if distanceSum >= distance {
			// Get new location by getting the distance between the previous and next waypoint
			return LocationFromLatLng(getNewCoordinates(previous.LatLng, waypoint.LatLng, distance-(distanceSum-diff)))
		}

		previous = waypoint
	}

	return LocationFromLatLng(previous.LatLng)
}
