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
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/monzo/typhon"
	"go.mongodb.org/mongo-driver/mongo"
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
		res.err = data
		switch data {
		case service.ErrResourceNotFound:
			res.code = http.StatusNotFound
		case service.ErrBadSyntax:
			res.code = http.StatusBadRequest
		case service.ErrForbidden:
			res.code = http.StatusForbidden
		case service.ErrResourceAlreadyExists:
			res.code = http.StatusConflict
		default:
			res.code = http.StatusInternalServerError
		}
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
	users      *users.Users
	challenges *challenges.Challenges
	activities *activities.Activities
}

func NewAPI(cfg Config) (*API, error) {

	db := NewDB(cfg.MongodbURI, cfg.DBName)
	activities := activities.NewActivities(db.Collection("activities"))
	users := users.NewUsers(db.Collection("users"), activities)
	challenges := challenges.NewChallenges(db.Collection("challenges"), users)

	return &API{
		typhon.Router{},
		cfg.Environment,
		cfg,
		db,
		users,
		challenges,
		activities,
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

		// if i := locations.Initialised(); !i {
		// 	return req.ResponseWithCode(nil, http.StatusServiceUnavailable)
		// }

		return req.ResponseWithCode(nil, http.StatusNoContent)
	})

	// Admin Routes
	a.GET("/api/admin/:userID", serve(a.GetAdmin, []typhon.Filter{}))       // admin
	a.PUT("/api/admin/:userID", serve(a.PutAdmin, []typhon.Filter{}))       // admin
	a.DELETE("/api/admin/:userID", serve(a.DeleteAdmin, []typhon.Filter{})) // admin

	// Challenges Routes
	a.GET("/api/challenges", serve(a.GetChallenges, []typhon.Filter{}))          // public
	a.POST("/api/challenges", serve(a.PostChallenge, []typhon.Filter{}))         // auth
	a.GET("/api/challenges/:id", serve(a.GetChallenge, []typhon.Filter{}))       // public
	a.DELETE("/api/challenges/:id", serve(a.DeleteChallenge, []typhon.Filter{})) // auth
	// a.PATCH("/api/challenges/:id", serve(a.PatchChallenge, []typhon.Filter{})) // auth
	a.GET("/api/challenges/:id/members/:userID/progress", serve(a.GetProgress, []typhon.Filter{})) // public

	// User routes
	a.GET("/api/users", serve(a.GetUsers, []typhon.Filter{}))        // admin
	a.GET("/api/users/:userID", serve(a.GetUser, []typhon.Filter{})) // auth
	// a.PATCH("/api/users/:userID", serve(a.PatchUser, []typhon.Filter{})) // valid user
	a.DELETE("/api/users/:userID", serve(a.DeleteUser, []typhon.Filter{})) // valid user
	a.PUT("/api/users/:userID", serve(a.PutUser, []typhon.Filter{}))       // valid user

	// User activities routes
	a.GET("/api/users/:userID/activities", serve(a.GetUserActivities, []typhon.Filter{}))           // public
	a.POST("/api/users/:userID/activities", serve(a.PostUserActivity, []typhon.Filter{}))           // valid user
	a.GET("/api/users/:userID/activities/:activityID", serve(a.GetUserActivity, []typhon.Filter{})) // public
	// a.PATCH("/api/users/:userID/activities/:activityID", serve(a.PatchUserActivity, []typhon.Filter{})) // valid user
	a.DELETE("/api/users/:userID/activities/:activityID", serve(a.DeleteUserActivity, []typhon.Filter{})) // valid user

	// User challenge routes
	a.PUT("/api/users/:userID/challenges/:id", serve(a.JoinChallenge, []typhon.Filter{}))     // valid user
	a.DELETE("/api/users/:userID/challenges/:id", serve(a.LeaveChallenge, []typhon.Filter{})) // valid user

	a.GET("/api/profile", serve(a.GetProfile, []typhon.Filter{})) // auth (maybe valid user?)

	// Make sure body filtering and logging go last!
	svc := a.Serve().
		Filter(typhon.H2cFilter).
		Filter(typhon.ErrorFilter).
		Filter(a.ActorFilter).
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
