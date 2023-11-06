package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

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

/*
* Test challenge endpoints
 */

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
