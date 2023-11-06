package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets/locations"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/monzo/typhon"
)

/*
* Test admin endpoints
 */

func TestGetAdmin(t *testing.T) {
	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodGet, fmt.Sprintf("%s/admin/test", baseURI), nil)
	res := req.Send().Response()

	if !admin && res.StatusCode == http.StatusOK {
		t.Fatalf("Allowed getting admin")
	}
}

func TestPutAdmin(t *testing.T) {
	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodPut, fmt.Sprintf("%s/admin/test2", baseURI), nil)
	res := req.Send().Response()

	if !admin && res.StatusCode == http.StatusNoContent {
		t.Fatalf("Allowed putting admin")
	}
}

func TestDeleteAdmin(t *testing.T) {
	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodDelete, fmt.Sprintf("%s/admin/test", baseURI), nil)
	res := req.Send().Response()

	if !admin && res.StatusCode == http.StatusNoContent {
		t.Fatalf("Allowed putting admin")
	}
}

/*
* Test challenge endpoints
 */

func TestPostChallenge(t *testing.T) {
	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodPost, fmt.Sprintf("%s/challenges", baseURI), challenges.Challenge{
		PartialChallengeWithTarget: challenges.PartialChallengeWithTarget{
			PartialChallenge: challenges.PartialChallenge{
				BaseChallenge: challenges.BaseChallenge{
					ID:          "test_challenge_2",
					Name:        "Test Challenge 2",
					Description: "This is a test challenge",

					StartDate:  time.Date(2023, 9, 30, 20, 17, 4, 225000000, time.UTC),
					EndDate:    time.Date(2023, 9, 30, 20, 18, 4, 225000000, time.UTC),
					Public:     false,
					InviteOnly: false,
				},
				CreatedBy: user,
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
	})
	res := req.Send().Response()

	if !hasUser() && res.StatusCode == http.StatusCreated {
		t.Fatalf("Allowed creating challenge")
	}
}

func TestDeleteChallenge(t *testing.T) {

	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodDelete, fmt.Sprintf("%s/challenges/test_challenge", baseURI), nil)
	res := req.Send().Response()

	if (!admin && user != "test3") && res.StatusCode == http.StatusNoContent {
		t.Fatalf("Allowed deleting challenge")
	}
}

func TestJoinChallenge(t *testing.T) {

	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodPut, fmt.Sprintf("%s/users/test2/challenges/test_challenge", baseURI), nil)
	res := req.Send().Response()

	if !admin && res.StatusCode == http.StatusNoContent {
		t.Fatalf("Allowed joining challenge")
	}
}

func TestLeaveChallenge(t *testing.T) {

	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodDelete, fmt.Sprintf("%s/users/test2/challenges/test_challenge", baseURI), nil)
	res := req.Send().Response()

	if !admin && res.StatusCode == http.StatusNoContent {
		t.Fatalf("Allowed leaving challenge")
	}
}

/*
* Test user endpoints
 */

func TestPostWrongUserActivities(t *testing.T) {
	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodPost, fmt.Sprintf("%s/users/test2/activities", baseURI), map[string]interface{}{
		"type":  "walking",
		"value": 6,
	})
	res := req.Send().Response()

	if !admin && res.StatusCode == http.StatusCreated {
		t.Fatalf("Activity created for wrong user")
	}
}

func TestGetUsers(t *testing.T) {
	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodGet, fmt.Sprintf("%s/users", baseURI), nil)
	res := req.Send().Response()

	if !admin && res.StatusCode == http.StatusOK {
		t.Fatalf("Allowed getting users")
	}
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodDelete, fmt.Sprintf("%s/users/test3", baseURI), nil)
	res := req.Send().Response()

	if !admin && res.StatusCode == http.StatusNoContent {
		t.Fatalf("Allowed deleting user")
	}
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodPut, fmt.Sprintf("%s/users/testuser", baseURI), users.User{
		PartialUser: users.PartialUser{
			ID:        "testuser",
			FirstName: "Test",
			LastName:  "User",
		},
		Bio: "This is a test user",
	})
	res := req.Send().Response()

	if !admin && res.StatusCode == http.StatusCreated {
		t.Fatalf("Allowed creating user")
	}
}
