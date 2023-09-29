package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/admin"
	"github.com/AustinBayley/activity_tracker_api/pkg/auth"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/locations"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/monzo/typhon"
	"go.mongodb.org/mongo-driver/mongo"
	"googlemaps.github.io/maps"
)

type API struct {
	typhon.Router
	cfg        Config
	db         *mongo.Database
	auth       *auth.Auth
	users      *users.Users
	challenges *challenges.Challenges
	activities *activities.Activities
	admin      *admin.Admin
}

func NewAPI(cfg Config) (*API, error) {

	db := NewDB(cfg.MongodbURI, cfg.DBName)
	users := users.NewUsers(db.Collection("users"))
	challenges := challenges.NewChallenges(db.Collection("challenges"))
	activities := activities.NewActivities(db.Collection("activities"))

	auth, err := auth.NewAuth(cfg.ProjectID)
	if err != nil {
		return nil, err
	}
	admin := admin.NewAdmin(auth)

	c, err := maps.NewClient(maps.WithAPIKey(cfg.MapsAPIKey))
	if err != nil {
		return nil, err
	}
	_ = locations.NewLocations(c)

	return &API{
		typhon.Router{},
		cfg,
		db,
		auth,
		users,
		challenges,
		activities,
		admin,
	}, nil
}

func addFilters(svc typhon.Service, filters []typhon.Filter) typhon.Service {
	for _, f := range filters {
		svc = svc.Filter(f)
	}
	return svc
}

func (a *API) Start() {

	// Get health of service
	a.GET("/health", func(req typhon.Request) typhon.Response {
		// Test Mongodb connection
		err := a.db.Client().Ping(req.Context, nil)
		if err != nil {
			return req.ResponseWithCode(err, http.StatusServiceUnavailable)
		}

		if i := locations.Initialised(); !i {
			return req.ResponseWithCode(nil, http.StatusServiceUnavailable)
		}

		return req.ResponseWithCode(nil, http.StatusNoContent)
	})

	// Admin Routes
	a.GET("/admin/:userID", addFilters(a.GetAdmin, []typhon.Filter{a.ValidUserFilter}))
	a.PUT("/admin/:userID", addFilters(a.PutAdmin, []typhon.Filter{a.ValidUserFilter}))
	a.DELETE("/admin/:userID", addFilters(a.DeleteAdmin, []typhon.Filter{a.ValidUserFilter}))

	// Challenges Routes
	a.GET("/challenges", addFilters(a.GetChallenges, []typhon.Filter{}))
	a.POST("/challenges", addFilters(a.PostChallenge, []typhon.Filter{}))
	a.GET("/challenges/:id", addFilters(a.GetChallenge, []typhon.Filter{}))
	a.DELETE("/challenges/:id", addFilters(a.DeleteChallenge, []typhon.Filter{}))
	a.PATCH("/challenges/:id", addFilters(a.PatchChallenge, []typhon.Filter{}))
	a.PUT("/challenges/:id/members/:userID", addFilters(a.PutMember, []typhon.Filter{a.ValidUserFilter}))
	a.DELETE("/challenges/:id/members/:userID", addFilters(a.DeleteMember, []typhon.Filter{a.ValidUserFilter}))

	// User routes
	a.GET("/users", addFilters(a.GetUsers, []typhon.Filter{}))
	a.GET("/users/:userID", addFilters(a.GetUser, []typhon.Filter{}))
	a.PATCH("/users/:userID", addFilters(a.PatchUser, []typhon.Filter{a.ValidUserFilter}))
	a.DELETE("/users/:userID", addFilters(a.DeleteUser, []typhon.Filter{a.ValidUserFilter}))
	// Because this method will only run once, a valid user filter is not required as it will not change other than via patch or delete requests
	a.PUT("/users/:userID", addFilters(a.PutUser, []typhon.Filter{}))
	a.GET("/users/:userID/activities", addFilters(a.GetUserActivities, []typhon.Filter{}))
	a.POST("/users/:userID/activities", addFilters(a.PostUserActivity, []typhon.Filter{a.ValidUserFilter}))
	a.GET("/users/:userID/activities/:activityID", addFilters(a.GetUserActivity, []typhon.Filter{}))
	a.PATCH("/users/:userID/activities/:activityID", addFilters(a.PatchUserActivity, []typhon.Filter{a.ValidUserFilter}))
	a.DELETE("/users/:userID/activities/:activityID", addFilters(a.DeleteUserActivity, []typhon.Filter{a.ValidUserFilter}))

	// Make sure body filtering and logging go last!
	svc := a.Serve().
		Filter(typhon.H2cFilter).
		Filter(typhon.ErrorFilter).
		Filter(a.BodyFilter).
		Filter(Logging)

	defer func() {
		log.Printf("Shutting down database connection")

		if err := a.db.Client().Disconnect(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}()

	srv, err := typhon.Listen(svc, fmt.Sprintf(":%d", a.cfg.Port), typhon.WithTimeout(typhon.TimeoutOptions{Read: time.Second * 10}))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("👋  Listening on %v", srv.Listener().Addr())

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	log.Printf("☠️  Shutting down")
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Stop(c)
}
