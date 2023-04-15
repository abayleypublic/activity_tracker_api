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
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/monzo/typhon"
)

type API struct {
	cfg        Config
	users      *users.Users
	challenges *challenges.Challenges
	activities *activities.Activities
}

func NewAPI(cfg Config) *API {

	db := NewDb(cfg)

	users := users.NewUsers(db.Collection("users"))
	challenges := challenges.NewChallenges(db.Collection("challenges"))
	activities := activities.NewActivities(db.Collection("activities"))

	return &API{
		cfg,
		users,
		challenges,
		activities,
	}
}

func Logging(req typhon.Request, svc typhon.Service) typhon.Response {
	log.Printf("📡 %v %v - %v", req.Method, req.URL, req.RemoteAddr)
	return svc(req)
}

func (a *API) Start() {

	r := typhon.Router{}

	r.GET("/health", func(req typhon.Request) typhon.Response {
		return req.Response("OK")
	})

	svc := r.Serve().
		Filter(typhon.ErrorFilter).
		Filter(typhon.H2cFilter).
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
