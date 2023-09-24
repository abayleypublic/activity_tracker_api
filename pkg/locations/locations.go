package locations

import (
	"context"
	"errors"

	"googlemaps.github.io/maps"
)

// Becuase the locations package requires a maps.Client to be initialized, this poses a challenge for use in other packages
// therefore this package is defined as a singleton.
var (
	instance *Locations
)

func Initialised() bool {
	return instance != nil
}

type Locations struct {
	*maps.Client
}

func NewLocations(c *maps.Client) *Locations {
	if instance == nil {
		instance = &Locations{c}
	}
	return instance
}

type LatLng maps.LatLng

func GetLocationName(ctx context.Context, latlng LatLng) (string, error) {

	if !Initialised() {
		return "", errors.New("locations instance not initialized")
	}

	mll := maps.LatLng(latlng)
	place, err := instance.ReverseGeocode(ctx, &maps.GeocodingRequest{
		LatLng:     &mll,
		ResultType: []string{"political"},
	})
	if err != nil {
		return "", err
	}

	return place[0].FormattedAddress, nil
}

type Location struct {
	LatLng LatLng `json:"latlng" bson:"latlng"`
	Name   string `json:"name" bson:"name"`
}

func LocationFromLatLng(ctx context.Context, latlng LatLng) (Location, error) {
	name, err := GetLocationName(ctx, latlng)
	if err != nil {
		return Location{}, err
	}

	return Location{
		LatLng: latlng,
		Name:   name,
	}, nil
}
