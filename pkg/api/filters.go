package api

import (
	"fmt"
	"log"

	"github.com/AustinBayley/activity_tracker_api/pkg/auth"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"
)

func Logging(req typhon.Request, svc typhon.Service) typhon.Response {
	log.Printf("📡 %v %v - %v", req.Method, req.URL, req.RemoteAddr)
	return svc(req)
}

func HasAuth(req typhon.Request, svc typhon.Service) typhon.Response {
	if req.Header.Get("Authorization") == "" {
		return req.Response("Unauthorized")
	}
	return svc(req)
}

func ValidAuth(req typhon.Request, svc typhon.Service) typhon.Response {

	t, err := auth.GetAuthToken(req)
	if err != nil {
		return req.Response(terrors.Unauthorized("", "error getting ID token", nil))
	}

	token, err := auth.GetValidToken(t)
	if err != nil {
		return req.Response(err)
	}

	fmt.Println(token)

	return svc(req)
}
