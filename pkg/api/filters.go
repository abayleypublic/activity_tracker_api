package api

import (
	"encoding/json"

	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"
)

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
		return a.Error(req, err)
	}

	_, err = a.auth.GetValidToken(req.Context, t)
	if err != nil {
		return a.Error(req, err)
	}

	return svc(req)
}

// Only allow admin users
func (a *API) AdminAuthFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	t, err := a.auth.GetAuthToken(req)
	if err != nil {
		return a.Error(req, err)
	}

	token, err := a.auth.GetValidToken(req.Context, t)
	if err != nil {
		return a.Error(req, err)
	}

	if admin := a.auth.IsAdmin(*token); !admin {
		return a.Error(req, terrors.Unauthorized("", "user is not admin", nil))
	}

	return svc(req)
}

// Check if userID is equal to token subject or token is admin
func (a *API) ValidUserFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "could not determine target user", nil))
	}
	userID := uuid.ID(id)

	t, err := a.auth.GetAuthToken(req)
	if err != nil {
		return a.Error(req, err)
	}

	token, err := a.auth.GetToken(req.Context, t)
	if err != nil {
		return a.Error(req, err)
	}

	tokenSubject := a.auth.GetUserID(req.Context, *token)

	if admin := a.auth.IsAdmin(*token); !admin && tokenSubject != userID {
		return a.Error(req, terrors.Unauthorized("", "user is not authorized to perform this action", nil))
	}

	return svc(req)
}

type PartialTerror struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a *API) BodyFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	res := svc(req)

	if res.Error != nil {
		if res.Body != nil {
			if b, err := res.BodyBytes(true); err == nil {
				var err PartialTerror
				json.Unmarshal(b, &err)
				res.Encode(err)
			}
			res.Body.Close()
		}
	}

	return res
}
