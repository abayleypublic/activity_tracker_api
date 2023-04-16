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
)

type API struct {
	typhon.Router
	cfg        Config
	auth       *auth.Auth
	users      *users.Users
	challenges *challenges.Challenges
	activities *activities.Activities
	admin      *admin.Admin
}

func NewAPI(cfg Config) (*API, error) {

	db := NewDb(cfg)

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

	a.GET("/health", func(req typhon.Request) typhon.Response {
		return req.Response("OK")
	})

	a.GET("/users", addFilters(a.GetUsers, []typhon.Filter{a.AdminAuthFilter}))

	a.GET("/admin/:id", a.GetAdmin)
	a.PUT("/admin/:id", a.PutAdmin)
	a.DELETE("/admin/:id", a.DeleteAdmin)

	// Make sure body filtering and logging go last!
	svc := a.Serve().
		Filter(typhon.H2cFilter).
		Filter(typhon.ErrorFilter).
		Filter(a.BodyFilter).
		Filter(Logging)

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
