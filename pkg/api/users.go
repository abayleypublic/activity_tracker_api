package api

import (
	"net/http"

	"github.com/monzo/terrors"
	"github.com/monzo/typhon"
)

func (a *API) GetUsers(req typhon.Request) typhon.Response {

	u, err := a.users.GetUsers(req.Context)
	if err != nil {
		return a.Error(req, err)
	}

	return req.Response(u)

}

func (a *API) GetUser(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	u, err := a.users.GetUser(req.Context, id)
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.Response(u)

}

func (a *API) PatchUser(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) DeleteUser(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	_, err := a.users.DeleteUser(req.Context, id)
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.ResponseWithCode(interface{}(nil), http.StatusOK)

}

func (a *API) PutUser(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) DownloadUserData(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) GetUserActivity(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) PostUserActivity(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) PatchUserActivity(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) DeleteUserActivity(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) GetUserActivities(req typhon.Request) typhon.Response {

	return req.Response("OK")

}
