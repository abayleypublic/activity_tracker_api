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
	"github.com/AustinBayley/activity_tracker_api/pkg/targets/locations"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/monzo/typhon"
	"go.mongodb.org/mongo-driver/mongo"
	"googlemaps.github.io/maps"
)

type Environment string

const (
	DEV  Environment = "dev"
	STG  Environment = "stg"
	PROD Environment = "prod"
)

type Response struct {
	Data interface{}
	code int   `json:"-"`
	err  error `json:"-"`
}

func NewResponseWithCode(data interface{}, code int) Response {
	res := Response{data, code, nil}

	switch data := data.(type) {
	case *Error:
		res.code = data.Code
		res.err = data
	case error:
		res.code = http.StatusInternalServerError
		res.err = data
	}

	return res
}

func NewResponse(data interface{}) Response {
	return NewResponseWithCode(data, http.StatusOK)
}

type Service func(req typhon.Request) Response

func addFilters(svc typhon.Service, filters []typhon.Filter) typhon.Service {
	for _, f := range filters {
		svc = svc.Filter(f)
	}
	return svc
}

func serve(service Service, filters []typhon.Filter) typhon.Service {
	return addFilters(func(req typhon.Request) typhon.Response {
		res := service(req)
		resp := req.ResponseWithCode(res.Data, res.code)
		resp.Error = res.err
		return resp
	}, filters)
}

type API struct {
	typhon.Router
	env        Environment
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
	challenges := challenges.NewChallenges(db.Collection("challenges"))
	activities := activities.NewActivities(db.Collection("activities"))
	users := users.NewUsers(db.Collection("users"), activities)

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
		cfg.Environment,
		cfg,
		db,
		auth,
		users,
		challenges,
		activities,
		admin,
	}, nil
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
	a.GET("/admin/:userID", serve(a.GetAdmin, []typhon.Filter{a.ValidUserFilter}))
	a.PUT("/admin/:userID", serve(a.PutAdmin, []typhon.Filter{a.ValidUserFilter}))
	a.DELETE("/admin/:userID", serve(a.DeleteAdmin, []typhon.Filter{a.ValidUserFilter}))

	// Challenges Routes
	a.GET("/challenges", serve(a.GetChallenges, []typhon.Filter{}))
	a.POST("/challenges", serve(a.PostChallenge, []typhon.Filter{}))
	a.GET("/challenges/:id", serve(a.GetChallenge, []typhon.Filter{}))
	a.DELETE("/challenges/:id", serve(a.DeleteChallenge, []typhon.Filter{}))
	a.PATCH("/challenges/:id", serve(a.PatchChallenge, []typhon.Filter{}))

	// User routes
	a.GET("/users", serve(a.GetUsers, []typhon.Filter{}))
	a.GET("/users/:userID", serve(a.GetUser, []typhon.Filter{}))
	a.PATCH("/users/:userID", serve(a.PatchUser, []typhon.Filter{a.ValidUserFilter}))
	a.DELETE("/users/:userID", serve(a.DeleteUser, []typhon.Filter{a.ValidUserFilter}))
	// Because this method will only run once, a valid user filter is not required as it will not change other than via patch or delete requests
	a.PUT("/users/:userID", serve(a.PutUser, []typhon.Filter{}))

	// User activities routes
	a.GET("/users/:userID/activities", serve(a.GetUserActivities, []typhon.Filter{}))
	a.POST("/users/:userID/activities", serve(a.PostUserActivity, []typhon.Filter{a.ValidUserFilter}))
	a.GET("/users/:userID/activities/:activityID", serve(a.GetUserActivity, []typhon.Filter{}))
	a.PATCH("/users/:userID/activities/:activityID", serve(a.PatchUserActivity, []typhon.Filter{a.ValidUserFilter}))
	a.DELETE("/users/:userID/activities/:activityID", serve(a.DeleteUserActivity, []typhon.Filter{a.ValidUserFilter}))

	// User challenge routes
	a.PUT("/users/:userID/challenges/:id", serve(a.JoinChallenge, []typhon.Filter{a.ValidUserFilter}))
	a.DELETE("/users/:userID/challenges/:id", serve(a.LeaveChallenge, []typhon.Filter{a.ValidUserFilter}))

	// Make sure body filtering and logging go last!
	svc := a.Serve().
		Filter(typhon.H2cFilter).
		Filter(typhon.ErrorFilter).
		// Filter(a.BodyFilter).
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
