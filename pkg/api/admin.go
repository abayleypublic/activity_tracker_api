package api

import (
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/errs"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"github.com/monzo/typhon"
)

func (a *API) GetAdmin(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "id not supplied")
	}

	admin, err := a.admin.GetAdmin(req.Context, uuid.ID(id))
	if err != nil {
		return errs.ForbiddenResponse(req, "user is not authorized to perform this action")
	}

	return req.Response(admin)

}

func (a *API) DeleteAdmin(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "id not supplied")
	}

	err := a.admin.SetAdmin(req.Context, uuid.ID(id), false)
	if err != nil {
		return errs.InternalServerResponse(req, "error deleting admin")
	}

	return req.ResponseWithCode(nil, http.StatusNoContent)

}

func (a *API) PutAdmin(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "id not supplied")
	}

	err := a.admin.SetAdmin(req.Context, uuid.ID(id), true)
	if err != nil {
		return errs.InternalServerResponse(req, "error setting admin")
	}

	return req.ResponseWithCode(nil, http.StatusNoContent)

}
