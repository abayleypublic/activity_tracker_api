package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/api"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/gin-gonic/gin"
)

func TestCreateUser(t *testing.T) {
	email := "testcreate@user.com"
	user := users.Detail{
		FirstName: "Test",
		LastName:  "User",
		Email:     email,
	}

	bb, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	req := httptest.NewRequest(
		"POST",
		"/users",
		strings.NewReader(string(bb)),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Request-Email", email)

	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.Request = req

	API.ActorFilter(ctx)
	timePre := time.Now()
	API.PostUser(ctx)

	if ctx.Writer.Status() != 201 {
		t.Fatalf("expected status 201, got %d", ctx.Writer.Status())
	}

	var createdUser users.Detail
	if err := json.NewDecoder(recorder.Body).Decode(&createdUser); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if createdUser.Email != email {
		t.Errorf("expected email %s, got %s", email, createdUser.Email)
	}

	if createdUser.FirstName != user.FirstName {
		t.Errorf("expected first name %s, got %s", user.FirstName, createdUser.FirstName)
	}

	if createdUser.LastName != user.LastName {
		t.Errorf("expected last name %s, got %s", user.LastName, createdUser.LastName)
	}

	if createdUser.CreatedDate.Unix() > timePre.Unix() {
		t.Errorf("expected created date to be after request was sent, got %s", createdUser.CreatedDate)
	}

	t.Cleanup(func() {
		_ = Users.Delete(ctx, createdUser.ID)
	})
}

func TestReadUser(t *testing.T) {
	email := "testread@user.com"
	user, callback, err := CreateTestUser(context.Background(), email)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		callback()
	})

	_ = httptest.NewRequest("GET", "/users/"+string(user.ID), nil)
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.AddParam("userID", string(user.ID))

	API.GetUser(ctx)

	if ctx.Writer.Status() != 200 {
		t.Fatalf("expected status 200, got %d", ctx.Writer.Status())
	}

	var gotUser api.PartialUser
	if err := json.NewDecoder(recorder.Body).Decode(&gotUser); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if gotUser.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, gotUser.ID)
	}

	if gotUser.FirstName != user.FirstName {
		t.Errorf("expected first name %s, got %s", user.FirstName, gotUser.FirstName)
	}

	if gotUser.LastName != user.LastName {
		t.Errorf("expected last name %s, got %s", user.LastName, gotUser.LastName)
	}
}

func TestGetProfile(t *testing.T) {
	email := "testread@user.com"
	user, callback, err := CreateTestUser(context.Background(), email)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		callback()
	})

	_ = httptest.NewRequest("GET", "/profile", nil)
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: user.ID,
	})

	API.GetProfile(ctx)

	if ctx.Writer.Status() != 200 {
		t.Fatalf("expected status 200, got %d", ctx.Writer.Status())
	}

	var gotUser users.User
	if err := json.NewDecoder(recorder.Body).Decode(&gotUser); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if gotUser.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, gotUser.ID)
	}

	if gotUser.FirstName != user.FirstName {
		t.Errorf("expected first name %s, got %s", user.FirstName, gotUser.FirstName)
	}

	if gotUser.LastName != user.LastName {
		t.Errorf("expected last name %s, got %s", user.LastName, gotUser.LastName)
	}
}

func TestUpdateUser(t *testing.T) {
	email := "testupdate@user.com"
	user, callback, _ := CreateTestUser(context.Background(), email)
	t.Cleanup(func() {
		callback()
	})

	patch := `[{"op":"replace","path":"/first_name","value":"Updated"}]`
	req := httptest.NewRequest("PATCH", "/users/"+string(user.ID), strings.NewReader(patch))
	req.Header.Set("Content-Type", "application/json-patch+json")
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.AddParam("userID", string(user.ID))
	ctx.Request = req

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: user.ID,
	})

	API.PatchUser(ctx)

	if ctx.Writer.Status() != 204 {
		t.Fatalf("expected status 204, got %d", ctx.Writer.Status())
	}

	// Verify update
	var updated users.Detail
	if err := Users.Get(ctx, user.ID, &updated); err != nil {
		t.Fatalf("failed to get updated user: %v", err)
	}
	if updated.FirstName != "Updated" {
		t.Errorf("expected first name 'Updated', got %s", updated.FirstName)
	}
}

func TestDeleteUser(t *testing.T) {
	email := "testdelete@user.com"
	user, callback, _ := CreateTestUser(context.Background(), email)
	t.Cleanup(func() {
		callback()
	})

	_ = httptest.NewRequest("DELETE", "/users/"+string(user.ID), nil)
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.Params = gin.Params{{Key: "userID", Value: string(user.ID)}}

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: user.ID,
	})

	API.DeleteUser(ctx)

	if ctx.Writer.Status() != 204 {
		t.Fatalf("expected status 204, got %d", ctx.Writer.Status())
	}

	// Verify deletion
	var deleted users.Detail
	err := Users.Get(ctx, user.ID, &deleted)
	if err == nil {
		t.Errorf("expected error getting deleted user, got none")
	}
}

func TestJoinChallenge(t *testing.T) {
	email := "testjoinchallenge@user.com"
	user, callback, err := CreateTestUser(context.Background(), email)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	t.Cleanup(func() {
		callback()
	})

	challenge, cleanup, err := CreateTestChallenge(context.Background(), "Test Challenge")
	if err != nil {
		t.Fatalf("failed to create test challenge: %v", err)
	}
	t.Cleanup(cleanup)

	req := httptest.NewRequest("POST", "/users/"+string(user.ID)+"/challenges/"+string(challenge.ID), nil)
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.AddParam("userID", string(user.ID))
	ctx.AddParam("id", string(challenge.ID))
	ctx.Request = req

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: user.ID,
	})

	API.SetChallengeMembership(true)(ctx)

	if ctx.Writer.Status() != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", ctx.Writer.Status())
	}

	var updatedChallenge challenges.Challenge
	if err := Challenges.Get(context.Background(), challenge.ID, &updatedChallenge); err != nil {
		t.Fatalf("failed to get updated challenge: %v", err)
	}

	found := false
	for _, member := range updatedChallenge.Members {
		if member == user.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected user %s to be a member of the challenge", user.ID)
	}
}

func TestLeaveChallenge(t *testing.T) {
	email := "testleavechallenge@user.com"
	user, callback, err := CreateTestUser(context.Background(), email)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	t.Cleanup(func() {
		callback()
	})

	challenge, cleanup, err := CreateTestChallenge(context.Background(), "Test Challenge")
	if err != nil {
		t.Fatalf("failed to create test challenge: %v", err)
	}
	t.Cleanup(cleanup)

	err = Challenges.Update(context.Background(), challenges.SetMemberOperation{
		User:      user.ID,
		Challenge: challenge.ID,
		Member:    true,
	})
	if err != nil {
		t.Fatalf("failed to add user to challenge: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/users/"+string(user.ID)+"/challenges/"+string(challenge.ID), nil)
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.AddParam("userID", string(user.ID))
	ctx.AddParam("id", string(challenge.ID))
	ctx.Request = req

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: user.ID,
	})

	API.SetChallengeMembership(false)(ctx)

	if ctx.Writer.Status() != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", ctx.Writer.Status())
	}

	var updatedChallenge challenges.Challenge
	if err := Challenges.Get(context.Background(), challenge.ID, &updatedChallenge); err != nil {
		t.Fatalf("failed to get updated challenge: %v", err)
	}

	for _, member := range updatedChallenge.Members {
		if member == user.ID {
			t.Errorf("expected user %s to no longer be a member of the challenge", user.ID)
		}
	}
}
