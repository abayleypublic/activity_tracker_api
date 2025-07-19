package integration

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/api"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets/locations"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	baseURI string
	dbName  string = "activity-tracker"
	admin   bool
	user    service.ID = service.UnknownUser
	db      *mongo.Database
	as      *activities.Service
	us      *users.Users
	cs      *challenges.Challenges
)

func hasUser() bool {
	return user != service.UnknownUser
}

// Build out the database with dummy data
func init() {
	baseURI = os.Getenv("API_URI")
	if u := os.Getenv("User"); u != "" {
		user = service.ID(u)
	}

	mongoURI := os.Getenv("MONGODB_URI")

	adm, err := strconv.ParseBool(os.Getenv("ADMIN"))
	if err != nil {
		log.Fatalln(err)
	}
	admin = adm

	db = api.NewDB(mongoURI, dbName)
	as = activities.NewActivities(db.Collection("activities"))
	us = users.NewUsers(db.Collection("users"), as)
	cs = challenges.NewChallenges(db.Collection("challenges"), us)

}

func setup() error {
	ctx := context.Background()

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

	// Create user test3
	third := users.User{
		PartialUser: users.PartialUser{
			ID:        "test3",
			FirstName: "Test",
			LastName:  "Three",
		},
		Bio: "This is the third test user",
	}
	us.Create(ctx, third)

	// Create challenge
	c := challenges.Challenge{
		PartialChallengeWithTarget: challenges.PartialChallengeWithTarget{
			PartialChallenge: challenges.PartialChallenge{
				BaseChallenge: challenges.BaseChallenge{
					ID:          "test_challenge",
					Name:        "Test Challenge",
					Description: "This is a test challenge",

					StartDate:  time.Date(2023, 9, 30, 20, 17, 4, 225000000, time.UTC),
					EndDate:    time.Date(2023, 9, 30, 20, 18, 4, 225000000, time.UTC),
					Public:     false,
					InviteOnly: false,
				},
				CreatedBy: "test3",
			},
			Target: &targets.RouteMovingTarget{
				BaseTarget: targets.BaseTarget{
					TargetType: targets.RouteMovingTargetType,
				},
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
	return nil
}

func tearDown() error {
	ctx := context.Background()
	db.Collection("activities").DeleteMany(ctx, bson.M{})
	db.Collection("users").DeleteMany(ctx, bson.M{})
	db.Collection("challenges").DeleteMany(ctx, bson.M{})
	return nil
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		os.Exit(1)
	}

	exitCode := m.Run()

	if err := tearDown(); err != nil {
		os.Exit(1)
	}

	os.Exit(exitCode)

}
