package main

import (
	"context"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/api"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Config struct {
	Environment  api.Environment `envconfig:"ENVIRONMENT" default:"dev"`
	Port         int             `envconfig:"PORT" default:"8080"`
	MongoURI     string          `envconfig:"MONGODB_URI" required:"true"`
	DatabaseName string          `envconfig:"DATABASE_NAME" default:"activities"`
	AdminGroup   string          `envconfig:"ADMIN_GROUP" default:"activity_admin"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to process environment variables")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to connect to MongoDB")
	}

	db := client.Database(cfg.DatabaseName)

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

	err = api.NewAPI(api.NewConfig(
		cfg.Environment,
		db,
		cfg.Port,
		cfg.AdminGroup,
		acts,
		cs,
		us,
	)).Start()

	if err != nil {
		log.Fatal().
			Err(err).
			Msg("API failure")
	}
}
