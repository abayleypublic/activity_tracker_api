package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/admin"
	"github.com/AustinBayley/activity_tracker_api/pkg/auth"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/monzo/typhon"
	"go.mongodb.org/mongo-driver/mongo"
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

	db := NewDB(cfg)

	auth, err := auth.NewAuth(cfg.ProjectID)
	if err != nil {
		return nil, err
	}

	admin := admin.NewAdmin(auth)
	users := users.NewUsers(db.Collection("users"))
	challenges := challenges.NewChallenges(db.Collection("challenges"))
	activities := activities.NewActivities(db.Collection("activities"))

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
		return req.Response("OK")
	})

	// Activity Routes
	a.GET("/activities", addFilters(a.GetActivities, []typhon.Filter{}))
	a.PUT("/challenges/:id", addFilters(a.PutActivity, []typhon.Filter{}))
	a.DELETE("/challenges/:id", addFilters(a.DeleteActivity, []typhon.Filter{}))

	// Admin Routes
	a.GET("/admin/:id", addFilters(a.GetAdmin, []typhon.Filter{}))
	a.PUT("/admin/:id", addFilters(a.PutAdmin, []typhon.Filter{}))
	a.DELETE("/admin/:id", addFilters(a.DeleteAdmin, []typhon.Filter{}))

	// Challenges Routes
	a.GET("/challenges", addFilters(a.GetChallenges, []typhon.Filter{}))
	a.POST("/challenges", addFilters(a.PostChallenge, []typhon.Filter{}))
	a.GET("/challenges/:id", addFilters(a.GetChallenge, []typhon.Filter{}))
	a.DELETE("/challenges/:id", addFilters(a.DeleteChallenge, []typhon.Filter{}))
	a.PATCH("/challenges/:id", addFilters(a.PatchChallenge, []typhon.Filter{}))
	a.GET("/challenges/:id/members", addFilters(a.GetMembers, []typhon.Filter{}))
	a.PUT("/challenges/:id/members/:userID", addFilters(a.PutMember, []typhon.Filter{}))
	a.DELETE("/challenges/:id/members/:userID", addFilters(a.DeleteMember, []typhon.Filter{}))

	// User routes
	a.GET("/users", addFilters(a.GetUsers, []typhon.Filter{}))
	a.GET("/users/:id", addFilters(a.GetUser, []typhon.Filter{}))
	a.PATCH("/users/:id", addFilters(a.PatchUser, []typhon.Filter{}))
	a.DELETE("/users/:id", addFilters(a.DeleteUser, []typhon.Filter{}))
	a.PUT("/users/:id", addFilters(a.PutUser, []typhon.Filter{}))
	a.GET("/users/:id/activities", addFilters(a.GetUserActivities, []typhon.Filter{}))
	a.POST("/users/:id/activities", addFilters(a.PostUserActivity, []typhon.Filter{}))
	a.GET("/users/:id/activities/:activityID", addFilters(a.GetUserActivity, []typhon.Filter{}))
	a.PATCH("/users/:id/activities/:activityID", addFilters(a.PatchUserActivity, []typhon.Filter{}))
	a.DELETE("/users/:id/activities/:activityID", addFilters(a.DeleteUserActivity, []typhon.Filter{}))

	// Make sure body filtering and logging go last!
	svc := a.Serve().
		Filter(typhon.H2cFilter).
		Filter(typhon.ErrorFilter).
		Filter(a.BodyFilter).
		Filter(Logging)

	defer func() {

		log.Printf("Shutting down database connection")

		if err := a.db.Client().Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	srv, err := typhon.Listen(svc, fmt.Sprintf(":%d", a.cfg.Port), typhon.WithTimeout(typhon.TimeoutOptions{Read: time.Second * 10}))
	if err != nil {
		panic(err)
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

func (a *API) Error(req typhon.Request, err error) typhon.Response {
	res := req.Response(err.Error())
	res.Error = err
	return res
}
