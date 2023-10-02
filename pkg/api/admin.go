package api

import (
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"github.com/monzo/typhon"
)

func (a *API) GetAdmin(req typhon.Request) Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	admin, err := a.admin.GetAdmin(req.Context, uuid.ID(id))
	if err != nil {
		return NewResponse(Forbidden("user is not authorized to perform this action", err))
	}

	return NewResponse(admin)

}

func (a *API) DeleteAdmin(req typhon.Request) Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	err := a.admin.SetAdmin(req.Context, uuid.ID(id), false)
	if err != nil {
		return NewResponse(InternalServer("error deleting admin", err))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)

}

func (a *API) PutAdmin(req typhon.Request) Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	err := a.admin.SetAdmin(req.Context, uuid.ID(id), true)
	if err != nil {
		return NewResponse(InternalServer("error setting admin", nil))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)

}
