package api

import (
	"encoding/json"

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

	if err := res.Error; err != nil {
		slog.Error(req.Context, "📡 %v %v - %v - %v - %v", req.Method, req.URL, req.RemoteAddr, res.StatusCode, res.Error.Error())
	} else {
		slog.Debug(req.Context, "📡 %v %v - %v - %v", req.Method, req.URL, req.RemoteAddr, res.StatusCode)
	}

	return res
}

// Only allow valid tokens
func (a *API) ValidAuthFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	t, err := a.auth.GetAuthToken(req)
	if err != nil {
		return ForbiddenResponse(req, err.Error(), err)
	}

	_, err = a.auth.GetValidToken(req.Context, t)
	if err != nil {
		return ForbiddenResponse(req, err.Error(), err)
	}

	return svc(req)
}

// Only allow admin users
func (a *API) AdminAuthFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	t, err := a.auth.GetAuthToken(req)
	if err != nil {
		return UnauthorizedResponse(req, err.Error(), err)
	}

	token, err := a.auth.GetValidToken(req.Context, t)
	if err != nil {
		return ForbiddenResponse(req, err.Error(), err)
	}

	if admin := a.auth.IsAdmin(*token); !admin {
		return ForbiddenResponse(req, "user is not authorized to perform this action", err)
	}

	return svc(req)
}

// Check if userID is equal to token subject or token is admin
func (a *API) ValidUserFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	if a.env == DEV {
		return svc(req)
	}

	id, ok := a.Params(req)["userID"]
	if !ok {
		return BadRequestResponse(req, "could not determine target user", nil)
	}
	userID := service.ID(id)

	t, err := a.auth.GetAuthToken(req)
	if err != nil {
		return UnauthorizedResponse(req, err.Error(), err)
	}

	token, err := a.auth.GetToken(req.Context, t)
	if err != nil {
		return ForbiddenResponse(req, err.Error(), err)
	}

	tokenSubject := a.auth.GetUserID(req.Context, *token)

	if admin := a.auth.IsAdmin(*token); !admin && tokenSubject != userID {
		return ForbiddenResponse(req, "user is not authorized to perform this action", err)
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
