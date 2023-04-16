package api

import (
	"encoding/json"

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

// Only allow requests with authoirzation header
// func HasAuth(req typhon.Request, svc typhon.Service) typhon.Response {
// 	if req.Header.Get("Authorization") == "" {
// 		return ErrorResponse(req, terrors.Unauthorized("", "Authorization header not populated", nil))
// 	}
// 	return svc(req)
// }

// Only allow valid tokens
func (a *API) ValidAuthFilter(req typhon.Request, svc typhon.Service) typhon.Response {

	t, err := a.auth.GetAuthToken(req)
	if err != nil {
		return a.Error(req, err)
	}

	_, err = a.auth.GetValidToken(t)
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

	token, err := a.auth.GetValidToken(t)
	if err != nil {
		return a.Error(req, err)
	}

	admin, ok := token.Claims["admin"]

	if !ok {
		return a.Error(req, terrors.Unauthorized("", "admin property undefined", nil))
	}

	if !admin.(bool) {
		return a.Error(req, terrors.Unauthorized("", "user is not admin", nil))
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
