package api

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/monzo/slog"
	"github.com/monzo/typhon"
)

const (
	ADMIN_GROUP = "roam_admin"
)

func response(req typhon.Request, err *Error) typhon.Response {
	return req.ResponseWithCode(err, err.Code)
}

func BadRequestResponse(req typhon.Request, cause string, err error) typhon.Response {
	return response(req, BadRequest(cause, err))
}

func UnauthorizedResponse(req typhon.Request, cause string, err error) typhon.Response {
	return response(req, Unauthorized(cause, err))
}

func ForbiddenResponse(req typhon.Request, cause string, err error) typhon.Response {
	return response(req, Forbidden(cause, err))
}

func Logging(req typhon.Request, svc typhon.Service) typhon.Response {

	if req.URL.Path == "/health" {
		return svc(req)
	}

	res := svc(req)
	user, err := service.GetActorContext(res.Request.Context)
	if err != nil {
		slog.Error(req.Context, "📡 %v %v - %v - %v - %v - %v - %v", req.Method, req.URL, req.RemoteAddr, user.UserID, user.Admin, res.StatusCode, "failed to get actor context")
		return res
	}

	if err := res.Error; err != nil {
		slog.Error(req.Context, "📡 %v %v - %v - %v - %v - %v - %v", req.Method, req.URL, req.RemoteAddr, user.UserID, user.Admin, res.StatusCode, res.Error.Error())
	} else {
		slog.Debug(req.Context, "📡 %v %v - %v - %v - %v - %v", req.Method, req.URL, req.RemoteAddr, user.UserID, user.Admin, res.StatusCode)
	}

	return res
}

// ActorFilter updates the context with details of the user making the request.
func (a *API) ActorFilter(req typhon.Request, svc typhon.Service) typhon.Response {
	email := req.Header.Get("X-Auth-Request-Email")
	groups := req.Header.Get("X-Auth-Request-Groups")

	if a.env == DEV {
		req.Context = context.WithValue(req.Context, service.UserCtxKey, a.cfg.UserContext)
		return svc(req)
	}

	req.Context = context.WithValue(req.Context, service.UserCtxKey, service.RequestContext{
		UserID: service.ID(email),
		Admin:  strings.Contains(groups, ADMIN_GROUP),
	})

	return svc(req)
}

// Only allow requests with tokens
func (a *API) HasAuthFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	reqCtx, err := service.GetActorContext(req.Context)
	if err != nil {
		return ForbiddenResponse(req, "user is not authorized to perform this action", err)
	}

	// If the user has not been set, return unauthorized
	if reqCtx.UserID == service.UnknownUser {
		return ForbiddenResponse(req, "user is not authorized to perform this action", nil)
	}

	return svc(req)
}

// Check if userID is equal to token subject or token is admin
func (a *API) ValidUserFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return BadRequestResponse(req, "could not determine target user", nil)
	}
	userID := service.ID(id)

	reqCtx, err := service.GetActorContext(req.Context)
	if err != nil {
		return ForbiddenResponse(req, "user is not authorized to perform this action", err)
	}

	if !reqCtx.Admin && reqCtx.UserID != userID {
		return ForbiddenResponse(req, "user is not authorized to perform this action", nil)
	}

	return svc(req)
}

// Only allow admin users
func (a *API) AdminAuthFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	reqCtx, err := service.GetActorContext(req.Context)
	if err != nil {
		return ForbiddenResponse(req, "user is not authorized to perform this action", err)
	}

	if !reqCtx.Admin {
		return ForbiddenResponse(req, "user is not authorized to perform this action", nil)
	}

	return svc(req)
}

func (a *API) BodyFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	res := svc(req)

	if res.Error != nil {
		if res.Body != nil {
			if b, err := res.BodyBytes(true); err == nil {
				var err error
				json.Unmarshal(b, &err)
				res.Encode(err)
			}
			res.Body.Close()
		}
	}

	return res
}
