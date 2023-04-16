package admin

import (
	"github.com/AustinBayley/activity_tracker_api/pkg/auth"
	"github.com/monzo/typhon"
)

type Admin struct {
	auth *auth.Auth
}

func NewAdmin(auth *auth.Auth) *Admin {
	return &Admin{
		auth,
	}
}

func (a *Admin) GetAdmin(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *Admin) DeleteAdmin(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *Admin) PutAdmin(req typhon.Request) typhon.Response {

	return req.Response("OK")

}
