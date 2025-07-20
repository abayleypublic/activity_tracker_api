package api_test

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/api"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	AdminGroup = "activity_admin"
)

var (
	API        *api.API
	Users      *users.Service
	Challenges *challenges.Service
	Activities *activities.Service
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to connect to MongoDB")
	}

	db := client.Database("activity_tracker_test")

	acts := activities.New(db.Collection("activities"))
	cds := challenges.NewDetails(db.Collection("challenges"))
	ms := challenges.NewMemberships(db.Collection("memberships"))
	cs := challenges.New(cds, ms)

	uds := users.NewDetails(db.Collection("users"))
	us := users.New(
		uds,
		ms,
		cs,
		acts,
	)

	ctx := context.Background()
	if err := cs.Setup(ctx); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to setup challenges service")
	}

	if err := acts.Setup(ctx); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to setup activities service")
	}

	if err := us.Setup(ctx); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to setup users service")
	}

	Users = us
	Challenges = cs
	Activities = acts

	API = api.NewAPI(api.NewConfig(
		api.STG,
		db,
		80,
		AdminGroup,
		acts,
		cs,
		us,
	))

	// Setup code if needed
	code := m.Run()

	if err := db.Drop(ctx); err != nil {
		log.Error().
			Err(err).
			Msg("failed to drop test database")
	}

	// Teardown code if needed
	os.Exit(code)
}

func TestHealth(t *testing.T) {
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), API.Engine)
	API.HealthCheck(ctx)

	if ctx.Writer.Status() != 200 {
		t.Errorf("expected status 200, got %d", ctx.Writer.Status())
	}

	t.Log("health check passed successfully")
}
