package api_test

import (
	"net/http/httptest"
	"testing"

	"github.com/AustinBayley/activity_tracker_api/pkg/api"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/gin-gonic/gin"
)

func TestActorFilterNoHeaders(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	API.ActorFilter(ctx)

	actor, ok := api.GetActorContext(ctx)
	if !ok {
		t.Fatal("expected actor context to be set, but it was not")
	}

	if actor.UserID != "" {
		t.Errorf("expected empty UserID, got %s", actor.UserID)
	}

	if actor.Admin {
		t.Error("expected Admin to be false, but it was true")
	}

	t.Log("UserID and Admin status are set correctly")
}

func TestActorFilterUnknownUserHeaders(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("X-Auth-Request-Email", "test@user.com")
	ctx.Request.Header.Set("X-Auth-Request-Groups", AdminGroup)

	API.ActorFilter(ctx)

	actor, ok := api.GetActorContext(ctx)
	if !ok {
		t.Fatal("expected actor context to be set, but it was not")
	}

	if actor.UserID != "" {
		t.Error("expected UserID to be empty")
	}

	if actor.Admin {
		t.Error("expected Admin to be false, but it was true")
	}

	t.Log("UserID and Admin status are set correctly")
}

func TestActorFilterValidUserHeaders(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	email := "test@user.com"
	ctx.Request.Header.Set("X-Auth-Request-Email", email)
	ctx.Request.Header.Set("X-Auth-Request-Groups", AdminGroup)

	API.ActorFilter(ctx)

	u := users.Detail{
		FirstName: "Test",
		LastName:  "User",
		Email:     email,
		Bio:       "I am a test user",
	}
	Users.Create(ctx, &u)

	actor, ok := api.GetActorContext(ctx)
	if !ok {
		t.Fatal("expected actor context to be set, but it was not")
	}

	if actor.UserID != "" {
		t.Error("expected UserID to be empty")
	}

	if actor.Admin {
		t.Error("expected Admin to be false, but it was true")
	}

	if err := Users.Delete(ctx, u.ID); err != nil {
		t.Errorf("failed to delete test user: %v", err)
	}

	t.Log("UserID and Admin status are set correctly")
}
