package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AustinBayley/activity_tracker_api/pkg/api"
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
}

func TestActorFilterNoHeadersUserExists(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	_, callback, err := CreateTestUser(ctx, "test@user.com")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	defer callback()

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
}

func TestActorFilterValidUserHeaders(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	email := "test@user.com"
	ctx.Request.Header.Set("X-Auth-Request-Email", email)
	ctx.Request.Header.Set("X-Auth-Request-Groups", AdminGroup)

	_, callback, err := CreateTestUser(ctx, email)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	defer callback()

	API.ActorFilter(ctx)

	actor, ok := api.GetActorContext(ctx)
	if !ok {
		t.Fatal("expected actor context to be set, but it was not")
	}

	if actor.UserID == "" {
		t.Error("UserID should not be empty")
	}

	if !actor.Admin {
		t.Error("expected Admin to be true")
	}

}

func TestHasAuthFilterNoContext(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	API.ActorFilter(ctx)

	API.HasAuthFilter(ctx)
	if ctx.Writer.Status() != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, ctx.Writer.Status())
	}
}

func TestHasAuthFilterWithContext(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	email := "test@user.com"
	ctx.Request.Header.Set("X-Auth-Request-Email", email)
	ctx.Request.Header.Set("X-Auth-Request-Groups", AdminGroup)

	_, callback, err := CreateTestUser(ctx, email)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	defer callback()

	API.ActorFilter(ctx)

	API.HasAuthFilter(ctx)
	if ctx.Writer.Status() != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
}

func TestAdminAuthFilterNoUserID(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	API.ActorFilter(ctx)

	API.AdminAuthFilter(ctx)
	if ctx.Writer.Status() != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, ctx.Writer.Status())
	}
}

func TestAdminAuthFilterValidUserID(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	email := "test@user.com"
	ctx.Request.Header.Set("X-Auth-Request-Email", email)
	ctx.Request.Header.Set("X-Auth-Request-Groups", AdminGroup)
	_, callback, err := CreateTestUser(ctx, email)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	defer callback()

	API.ActorFilter(ctx)
	API.AdminAuthFilter(ctx)
	if ctx.Writer.Status() != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
}

func TestAdminAuthFilterValidUserIDNotAdmin(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	email := "test@user.com"
	ctx.Request.Header.Set("X-Auth-Request-Email", email)
	ctx.Request.Header.Set("X-Auth-Request-Groups", "some-other-group")
	_, callback, err := CreateTestUser(ctx, email)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	defer callback()

	API.ActorFilter(ctx)
	API.AdminAuthFilter(ctx)
	if ctx.Writer.Status() != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, ctx.Writer.Status())
	}
}
