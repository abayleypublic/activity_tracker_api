package api

import (
	"net/http"

	"github.com/monzo/terrors"
	"github.com/monzo/typhon"
)

func (a *API) GetAdmin(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]

	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	admin, err := a.admin.GetAdmin(req.Context, id)

	if err != nil {
		return a.Error(req, terrors.BadRequest("", "error getting admin status", nil))
	}

	return req.Response(admin)

}

func (a *API) DeleteAdmin(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]

	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	a.admin.SetAdmin(req.Context, id, false)

	return req.ResponseWithCode(nil, http.StatusNoContent)

}

func (a *API) PutAdmin(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]

	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	a.admin.SetAdmin(req.Context, id, true)

	return req.ResponseWithCode(nil, http.StatusNoContent)

}
