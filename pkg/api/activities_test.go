package api_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/api"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/gin-gonic/gin"
)

func TestCreateActivity(t *testing.T) {
	activity := activities.Activity{
		Type:  activities.Running,
		Value: 10,
		Start: time.Now().Add(-2 * time.Hour),
		End:   time.Now().Add(-1 * time.Hour),
	}

	bb, err := json.Marshal(activity)
	if err != nil {
		t.Fatalf("failed to marshal activity: %v", err)
	}

	req := httptest.NewRequest(
		"POST",
		"/users/test_user/activities",
		strings.NewReader(string(bb)),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.Request = req
	ctx.AddParam("userID", "test_user")

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: service.ID("test_user"),
	})

	timePre := time.Now()
	API.PostUserActivity(ctx)

	if ctx.Writer.Status() != 201 {
		t.Fatalf("expected status 201, got %d", ctx.Writer.Status())
	}

	var createdActivity activities.Activity
	if err := json.NewDecoder(recorder.Body).Decode(&createdActivity); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if createdActivity.CreatedDate.Unix() < timePre.Unix() {
		t.Errorf("expected created date to be after request was sent, got %s", createdActivity.CreatedDate)
	}

	t.Cleanup(func() {
		_ = Activities.Delete(ctx, activities.ActivityDeleteOpts{ID: &createdActivity.ID})
	})
}

func TestReadActivity(t *testing.T) {
	activity, cleanup, err := CreateTestActivity(context.Background(), "Read Activity")
	if err != nil {
		t.Fatalf("failed to create test activity: %v", err)
	}
	t.Cleanup(cleanup)

	_ = httptest.NewRequest("GET", "/activities/"+string(activity.ID), nil)
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.AddParam("activityID", string(activity.ID))

	API.GetActivity(ctx)

	if ctx.Writer.Status() != 200 {
		t.Fatalf("expected status 200, got %d", ctx.Writer.Status())
	}

	var gotActivity activities.Activity
	if err := json.NewDecoder(recorder.Body).Decode(&gotActivity); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if gotActivity.ID != activity.ID {
		t.Errorf("expected activity ID %s, got %s", activity.ID, gotActivity.ID)
	}
}

func TestUpdateActivity(t *testing.T) {
	activity, cleanup, _ := CreateTestActivity(context.Background(), "Update Activity")
	t.Cleanup(cleanup)

	patch := `[{"op":"replace","path":"/value","value":20}]`
	req := httptest.NewRequest("PATCH", "/activities/"+string(activity.ID), strings.NewReader(patch))
	req.Header.Set("Content-Type", "application/json-patch+json")
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.AddParam("activityID", string(activity.ID))
	ctx.Request = req

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: service.ID("test_user"),
	})

	API.PatchActivity(ctx)

	if ctx.Writer.Status() != 204 {
		t.Fatalf("expected status 204, got %d", ctx.Writer.Status())
	}

	// Verify update
	var updated activities.Activity
	if err := Activities.Get(ctx, activity.ID, &updated); err != nil {
		t.Fatalf("failed to get updated activity: %v", err)
	}
}

func TestDeleteActivity(t *testing.T) {
	activity, cleanup, _ := CreateTestActivity(context.Background(), "Delete Activity")
	t.Cleanup(cleanup)

	_ = httptest.NewRequest("DELETE", "/activities/"+string(activity.ID), nil)
	recorder := httptest.NewRecorder()
	ctx := gin.CreateTestContextOnly(recorder, API.Engine)
	ctx.Params = gin.Params{{Key: "activityID", Value: string(activity.ID)}}

	ctx.Set(api.UserCtxKey, api.RequestContext{
		UserID: service.ID("test_user"),
	})

	API.DeleteActivity(ctx)

	if ctx.Writer.Status() != 204 {
		t.Fatalf("expected status 204, got %d", ctx.Writer.Status())
	}

	// Verify deletion
	var deleted activities.Activity
	err := Activities.Get(ctx, activity.ID, &deleted)
	if err == nil {
		t.Errorf("expected error getting deleted activity, got none")
	}
}
