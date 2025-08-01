package api_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/api"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
	"github.com/gin-gonic/gin"
)

func TestCreateChallenge(t *testing.T) {
	email := "testcreatechallenge@user.com"
	_, callback, err := CreateTestUser(context.Background(), email)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	challenge := challenges.Challenge{
		Detail: challenges.Detail{
			BaseDetail: challenges.BaseDetail{
				Name:        "Test Challenge",
				Description: "A test challenge",
				StartDate:   time.Now().Add(-20 * time.Hour),
				EndDate:     time.Now().Add(-18 * time.Hour),
			},
			Target: &targets.RouteMovingTarget{
				BaseTarget: targets.BaseTarget{
					TargetType: targets.RouteMovingTargetType,
				},
			},
		},
	}

	bb, err := json.Marshal(challenge)
	if err != nil {
		t.Fatalf("failed to marshal challenge: %v", err)
	}

	req := httptest.NewRequest(
		"POST",
		"/challenges",
		strings.NewReader(string(bb)),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Request-Email", email)

	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.Request = req

	API.ActorFilter(ctx)
	timePre := time.Now()
	API.PostChallenge(ctx)

	if ctx.Writer.Status() != 201 {
		t.Fatalf("expected status 201, got %d", ctx.Writer.Status())
	}

	var createdChallenge challenges.Challenge
	if err := json.NewDecoder(recorder.Body).Decode(&createdChallenge); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if createdChallenge.CreatedDate.Unix() < timePre.Unix() {
		t.Errorf("expected created date to be after request was sent, got %s", createdChallenge.CreatedDate)
	}

	t.Cleanup(func() {
		_ = callback()
		_ = Challenges.Delete(ctx, createdChallenge.ID)
	})
}

func TestReadChallenge(t *testing.T) {
	title := "Read Challenge"
	challenge, cleanup, err := CreateTestChallenge(context.Background(), title)
	if err != nil {
		t.Fatalf("failed to create test challenge: %v", err)
	}
	t.Cleanup(cleanup)

	req := httptest.NewRequest("GET", "/challenges/"+string(challenge.ID), nil)
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.AddParam("id", string(challenge.ID))
	ctx.Request = req

	API.GetChallenge(ctx)

	if ctx.Writer.Status() != 200 {
		t.Fatalf("expected status 200, got %d", ctx.Writer.Status())
	}

	var gotChallenge challenges.Challenge
	if err := json.NewDecoder(recorder.Body).Decode(&gotChallenge); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if gotChallenge.ID != challenge.ID {
		t.Errorf("expected challenge ID %s, got %s", challenge.ID, gotChallenge.ID)
	}
}

func TestUpdateChallenge(t *testing.T) {
	title := "Update Challenge"
	challenge, cleanup, _ := CreateTestChallenge(context.Background(), title)
	t.Cleanup(cleanup)

	patch := `[{"op":"replace","path":"/name","value":"Updated Challenge"}]`
	req := httptest.NewRequest("PATCH", "/challenges/"+string(challenge.ID), strings.NewReader(patch))
	req.Header.Set("Content-Type", "application/json-patch+json")
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.AddParam("id", string(challenge.ID))
	ctx.Request = req

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: service.ID("test_user"),
	})

	API.PatchChallenge(ctx)

	if ctx.Writer.Status() != 204 {
		t.Fatalf("expected status 204, got %d", ctx.Writer.Status())
	}

	// Verify update
	var updated challenges.Challenge
	if err := Challenges.Get(ctx, challenge.ID, &updated); err != nil {
		t.Fatalf("failed to get updated challenge: %v", err)
	}

	if updated.Detail.Name != "Updated Challenge" {
		t.Errorf("expected challenge name 'Updated Challenge', got '%s'", updated.Detail.Name)
	}
}

func TestDeleteChallenge(t *testing.T) {
	title := "Delete Challenge"
	challenge, cleanup, _ := CreateTestChallenge(context.Background(), title)
	t.Cleanup(cleanup)

	req := httptest.NewRequest("DELETE", "/challenges/"+string(challenge.ID), nil)
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.Params = gin.Params{{Key: "id", Value: string(challenge.ID)}}
	ctx.Request = req

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: service.ID("test_user"),
	})

	API.DeleteChallenge(ctx)

	if ctx.Writer.Status() != 204 {
		t.Fatalf("expected status 204, got %d", ctx.Writer.Status())
	}

	// Verify deletion
	var deleted challenges.Challenge
	err := Challenges.Get(ctx, challenge.ID, &deleted)
	if err == nil {
		t.Errorf("expected error getting deleted challenge, got none")
	}
}
