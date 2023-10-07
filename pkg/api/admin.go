package api

import (
	"log"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/monzo/typhon"
)

func (a *API) GetAdmin(req typhon.Request) Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	admin, err := a.admin.GetAdmin(req.Context, service.ID(id))
	if err != nil {
		return NewResponse(Forbidden("user is not authorized to perform this action", err))
	}

	return NewResponse(admin)

}

func (a *API) DeleteAdmin(req typhon.Request) Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	if err := a.admin.SetAdmin(req.Context, service.ID(id), false); err != nil {
		return NewResponse(InternalServer("error deleting admin", err))
	}

	if err := a.admin.RevokeTokens(req.Context, service.ID(id)); err != nil {
		return NewResponse(InternalServer("error revoking tokens", err))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)

}

func (a *API) PutAdmin(req typhon.Request) Response {

	log.Println("Putting admin")
	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	err := a.admin.SetAdmin(req.Context, service.ID(id), true)
	if err != nil {
		return NewResponse(InternalServer("error setting admin", nil))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)

}
