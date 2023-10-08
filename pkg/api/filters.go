package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/monzo/slog"
	"github.com/monzo/typhon"
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

	if a.env == DEV {
		req.Context = context.WithValue(req.Context, service.UserCtxKey, a.cfg.UserContext)
		return svc(req)
	}

	// Get token from headers
	t, err := a.auth.GetAuthToken(req)
	if err != nil {
		// If no token has been supplied, proceed with unknown user permissions
		req.Context = context.WithValue(req.Context, service.UserCtxKey, service.RequestContext{
			UserID: service.UnknownUser,
			Admin:  false,
		})
		return svc(req)
	}

	// If token was supplied, get user details
	token, err := a.auth.GetToken(req.Context, t)
	if err != nil {
		return ForbiddenResponse(req, "invalid token", err)
	}
	tokenSubject := a.auth.GetUserID(req.Context, *token)

	// Check token for admin claim
	admin := a.auth.IsAdmin(*token)

	// If admin claim or making anything other than a GET request, always check token is valid & not revoked
	if admin || req.Method != http.MethodGet || req.URL.Path == "/profile" {
		_, err = a.auth.GetValidToken(req.Context, t)
		if err != nil {
			return ForbiddenResponse(req, "invalid token", err)
		}
	}

	req.Context = context.WithValue(req.Context, service.UserCtxKey, service.RequestContext{
		UserID: tokenSubject,
		Admin:  admin,
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
