package api

import (
	"encoding/json"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	jsonpatch "github.com/evanphx/json-patch/v5"
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

// I have no idea whether this will work
func (a *API) PatchUser(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	bb, err := req.BodyBytes(true)

	su, err := a.users.GetUser(req.Context, id)
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	subb, err := json.Marshal(su)
	if err != nil {
		return a.Error(req, terrors.NotFound("", "error marshalling stored user", nil))
	}

	b, err := jsonpatch.CreateMergePatch(subb, bb)
	newUser, err := jsonpatch.CreateMergePatch(subb, b)

	var user users.User
	if err = json.Unmarshal(newUser, &user); err != nil {
		return a.Error(req, terrors.BadRequest("", "error unmarshalling user", nil))
	}

	_, err = a.users.PutUser(req.Context, user)
	if err != nil {
		return a.Error(req, terrors.BadRequest("", "error decoding user", nil))
	}

	if err != nil {
		return a.Error(req, terrors.BadRequest("", err.Error(), nil))
	}

	return req.ResponseWithCode(nil, http.StatusNoContent)

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

	return req.ResponseWithCode(nil, http.StatusOK)

}

func (a *API) PutUser(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	var user users.User
	err := req.Decode(user)
	if err != nil {
		return a.Error(req, terrors.BadRequest("", "error decoding user", nil))
	}

	if id != user.ID {
		return a.Error(req, terrors.BadRequest("", "user ID does not equal path ID", nil))
	}

	_, err = a.users.PutUser(req.Context, user)

	if err != nil {
		return a.Error(req, terrors.BadRequest("", err.Error(), nil))
	}

	return req.ResponseWithCode(nil, http.StatusNoContent)

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
