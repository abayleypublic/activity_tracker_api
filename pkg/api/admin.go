package api

import (
	"fmt"
	"net/http"

	"github.com/monzo/terrors"
	"github.com/monzo/typhon"
)

func (a *API) GetAdmin(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) DeleteAdmin(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) PutAdmin(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]

	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	client, err := a.auth.Auth(req.Context)
	if err != nil {
		return a.Error(req, terrors.BadRequest("", "error getting auth client", nil))
	}

	claims := map[string]interface{}{"admin": true}
	fmt.Println("Got here")
	if err = client.SetCustomUserClaims(req.Context, id, claims); err != nil {
		return a.Error(req, terrors.BadRequest("", "error setting custom claims", nil))
	}
	fmt.Println("Got here")

	return req.ResponseWithCode(nil, http.StatusCreated)

}
