package test

import (
	"context"
	"os"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/api"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets/locations"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
)

var (
	baseURI string
	dbName  string = "activity-tracker"
)

// Build out the database with dummy data
func init() {
	baseURI = os.Getenv("API_URI")
	mongoURI := os.Getenv("MONGODB_URI")
	ctx := context.Background()

	db := api.NewDB(mongoURI, dbName)
	as := activities.NewActivities(db.Collection("activities"))
	us := users.NewUsers(db.Collection("users"), as)
	cs := challenges.NewChallenges(db.Collection("challenges"), us)

	// Create user test - this is the user that we will impersonate
	first := users.User{
		PartialUser: users.PartialUser{
			ID:        "test",
			FirstName: "Test",
			LastName:  "One",
		},
		Bio: "This is the first test user",
	}
	us.Create(ctx, first)

	// Create user test2
	second := users.User{
		PartialUser: users.PartialUser{
			ID:        "test2",
			FirstName: "Test",
			LastName:  "Two",
		},
		Bio: "This is the second test user",
	}
	us.Create(ctx, second)

	// Create challenge
	c := challenges.Challenge{
		PartialChallengeWithTarget: challenges.PartialChallengeWithTarget{
			PartialChallenge: challenges.PartialChallenge{
				BaseChallenge: challenges.BaseChallenge{
					ID:          "Challenge1",
					Name:        "Test Challenge",
					Description: "This is a test challenge",

					StartDate:  time.Date(2023, 9, 30, 20, 17, 4, 225000000, time.UTC),
					EndDate:    time.Date(2023, 9, 30, 20, 18, 4, 225000000, time.UTC),
					Public:     false,
					InviteOnly: false,
				},
				CreatedBy: "test",
			},
			Target: &targets.RouteMovingTarget{
				Route: targets.Route{
					Waypoints: []locations.Waypoint{
						{
							LatLng: locations.LatLng{
								Lat: 51.5014708012926,
								Lng: -0.14184707849440084,
							},
						},
						{
							LatLng: locations.LatLng{
								Lat: 41.891031639230576,
								Lng: 12.492352743595536,
							},
						},
					},
				},
			},
		},
	}
	cs.Create(ctx, c)
}
