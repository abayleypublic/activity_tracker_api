package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/engine"
	"github.com/monzo/typhon"
)

type API struct {
	*engine.Engine
}

func NewAPI(e *engine.Engine) *API {
	return &API{
		e,
	}
}

func (a *API) Start() {

	r := typhon.Router{}

	r.GET("/health", func(req typhon.Request) typhon.Response {
		return req.Response("OK")
	})

	svc := r.Serve().
		Filter(typhon.ErrorFilter).
		Filter(typhon.H2cFilter)
	srv, err := typhon.Listen(svc, fmt.Sprintf(":%d", a.Engine.Config.Port), typhon.WithTimeout(typhon.TimeoutOptions{Read: time.Second * 10}))
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
