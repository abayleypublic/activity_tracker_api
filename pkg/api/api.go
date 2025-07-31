package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Environment string

const (
	DEV  Environment = "dev"
	STG  Environment = "stg"
	PROD Environment = "prod"
)

type ListOptions struct {
	Max  int64 `form:"max,default=10"`
	Page int64 `form:"page,default=1"`
}

type Config struct {
	Environment Environment
	Database    *mongo.Database
	Port        int
	AdminGroup  string

	// Services
	activities *activities.Service
	challenges *challenges.Service
	users      *users.Service
}

func NewConfig(
	environment Environment,
	database *mongo.Database,
	port int,
	adminGroup string,

	activities *activities.Service,
	challenges *challenges.Service,
	users *users.Service,
) Config {
	return Config{
		Environment: environment,
		Database:    database,
		Port:        port,
		AdminGroup:  adminGroup,
		activities:  activities,
		challenges:  challenges,
		users:       users,
	}
}

type API struct {
	*gin.Engine
	env        Environment
	port       int
	adminGroup string
	db         *mongo.Database
	users      *users.Service
	challenges *challenges.Service
	activities *activities.Service
}

func NewAPI(cfg Config) *API {
	return &API{
		gin.New(),
		cfg.Environment,
		cfg.Port,
		cfg.AdminGroup,
		cfg.Database,
		cfg.users,
		cfg.challenges,
		cfg.activities,
	}
}

func (a *API) Start() error {
	if a.env != DEV {
		gin.SetMode(gin.ReleaseMode)
	}

	a.Use(gin.Recovery())
	a.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{
			"/health",
		},
	}))
	a.Use(a.ActorFilter)

	// Get health of service
	a.GET("/health", a.HealthCheck)

	// Activity routes
	a.GET("/activities/:activityID", a.GetActivity)       // public
	a.PATCH("/activities/:activityID", a.PatchActivity)   // valid user
	a.DELETE("/activities/:activityID", a.DeleteActivity) // valid user

	// Challenge Routes
	a.GET("/challenges", a.GetChallenges)                            // public
	a.POST("/challenges", a.PostChallenge)                           // auth
	a.GET("/challenges/:id", a.GetChallenge)                         // public
	a.DELETE("/challenges/:id", a.DeleteChallenge)                   // auth
	a.PATCH("/challenges/:id", a.PatchChallenge)                     // auth
	a.GET("/challenges/:id/members/:userID/progress", a.GetProgress) // public

	// User routes
	a.GET("/users", a.AdminAuthFilter, a.GetUsers) // admin
	a.GET("/users/:userID", a.GetUser)             // auth
	a.PATCH("/users/:userID", a.PatchUser)         // valid user
	a.DELETE("/users/:userID", a.DeleteUser)       // valid user
	a.POST("/users", a.PostUser)                   // valid user

	// User activities routes
	a.POST("/users/:userID/activities", a.PostUserActivity) // valid user
	a.GET("/users/:userID/activities", a.GetUserActivities) // public

	// User challenge routes
	a.PUT("/users/:userID/challenges/:id", a.SetChallengeMembership(true))     // valid user
	a.DELETE("/users/:userID/challenges/:id", a.SetChallengeMembership(false)) // valid user

	a.GET("/profile", a.GetProfile) // auth (maybe valid user?)

	defer func() {
		log.Warn().
			Msg("shutting down database connection")

		if err := a.db.Client().Disconnect(context.Background()); err != nil {
			log.Error().
				Err(err).
				Msg("failed to disconnect from database")
		}
	}()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: a.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().
				Err(err).
				Msg("failed to start server")
		}

		log.Info().
			Str("address", srv.Addr).
			Msg("👋  server listening")

	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	log.Warn().
		Msg("☠️  shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Warn().
			Err(err).
			Msg("failed to gracefully shutdown server")
	}

	log.Info().
		Msg("👋  server shutdown")

	return nil
}

func (a *API) HealthCheck(req *gin.Context) {
	// Test Mongodb connection
	err := a.db.Client().Ping(req, nil)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to ping database")

		req.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Cause: "database connection failed",
		})
		return
	}

	req.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
