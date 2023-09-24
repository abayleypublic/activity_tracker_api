package api

import (
	"encoding/json"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"
)

func (a *API) GetUsers(req typhon.Request) typhon.Response {

	u, err := a.users.ReadUsers(req.Context)
	if err != nil {
		return a.Error(req, err)
	}

	return req.Response(u)

}

func (a *API) GetUser(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	u, err := a.users.ReadUser(req.Context, uuid.ID(id))
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.Response(u)

}

// PatchUser updates an existing user with the given ID using a JSON merge patch.
// The request body should contain a JSON merge patch that describes the changes to be made to the user.
// Returns a 400 Bad Request error if the ID is not supplied or if there is an error decoding the user.
// Returns a 404 Not Found error if the user with the given ID does not exist or if there is an error marshalling or unmarshalling the user.
// Returns a 204 No Content response if the user is successfully updated.
func (a *API) PatchUser(req typhon.Request) typhon.Response {

	// Get user ID
	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	userID := uuid.ID(id)

	// Get body & store as slice of bytes
	bb, err := req.BodyBytes(true)
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	// Stored user
	su, err := a.users.ReadUser(req.Context, userID)
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	// Stored user as slice of bytes
	subb, err := json.Marshal(su)
	if err != nil {
		return a.Error(req, terrors.NotFound("", "error marshalling stored user", nil))
	}

	// Decode requested patch
	patch, err := jsonpatch.DecodePatch(bb)
	if err != nil {
		return a.Error(req, terrors.BadRequest("", "could not decode request", nil))
	}

	// Apply patch to stored user to get modified document
	modified, err := patch.Apply(subb)
	if err != nil {
		return a.Error(req, terrors.BadResponse("", "could not apply patch", nil))
	}

	// Unmarshal modified document into user struct
	var user users.User
	if err = json.Unmarshal(modified, &user); err != nil {
		return a.Error(req, terrors.BadRequest("", "error unmarshalling user", nil))
	}

	// Update user
	if err = a.users.UpdateUser(req.Context, user); err != nil {
		return a.Error(req, terrors.BadRequest("", "error decoding user", nil))
	}

	return req.Response(user)

}

func (a *API) DeleteUser(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	if _, err := a.users.DeleteUser(req.Context, uuid.ID(id)); err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.ResponseWithCode(nil, http.StatusOK)

}

func (a *API) PutUser(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	userID := uuid.ID(id)

	var user users.User
	if err := req.Decode(&user); err != nil {
		return a.Error(req, terrors.BadRequest("", "error decoding user", nil))
	}

	if userID != user.ID {
		return a.Error(req, terrors.BadRequest("", "user ID does not equal path ID", nil))
	}

	if err := a.users.CreateUser(req.Context, user); err != nil {
		return a.Error(req, terrors.BadRequest("", err.Error(), nil))
	}

	return req.ResponseWithCode(nil, http.StatusNoContent)

}

func (a *API) DownloadUserData(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) GetUserActivity(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "user ID not supplied", nil))
	}

	aid, ok := a.Params(req)["activityID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "activity ID not supplied", nil))
	}

	userID := uuid.ID(id)
	activityID := uuid.ID(aid)

	res, err := a.users.ReadUserActivity(req.Context, userID, activityID)
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.Response(res)

}

func (a *API) PostUserActivity(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	userID := uuid.ID(id)

	var activity activities.Activity
	if err := req.Decode(&activity); err != nil {
		return a.Error(req, terrors.BadRequest("", "error decoding activity", nil))
	}

	res, err := a.users.CreateUserActivity(req.Context, userID, activity)
	if err != nil {
		return a.Error(req, terrors.BadRequest("", err.Error(), nil))
	}

	return req.Response(res)

}

func (a *API) PatchUserActivity(req typhon.Request) typhon.Response {

	// Get user ID
	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	// Get activity ID
	aid, ok := a.Params(req)["activityID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "activity ID not supplied", nil))
	}

	userID := uuid.ID(id)
	activityID := uuid.ID(aid)

	// Get body & store as slice of bytes
	bb, err := req.BodyBytes(true)
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	// Stored activity
	sa, err := a.users.ReadUserActivity(req.Context, userID, activityID)
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	// Stored activity as slice of bytes
	sabb, err := json.Marshal(sa)
	if err != nil {
		return a.Error(req, terrors.NotFound("", "error marshalling stored user", nil))
	}

	// Decode requested patch
	patch, err := jsonpatch.DecodePatch(bb)
	if err != nil {
		return a.Error(req, terrors.BadRequest("", "could not decode request", nil))
	}

	// Apply patch to stored activity to get modified document
	modified, err := patch.Apply(sabb)
	if err != nil {
		return a.Error(req, terrors.BadResponse("", "could not apply patch", nil))
	}

	// Unmarshal modified document into user struct
	var activity activities.Activity
	if err = json.Unmarshal(modified, &activity); err != nil {
		return a.Error(req, terrors.BadRequest("", "error unmarshalling activity", nil))
	}

	// Update activity
	res, err := a.users.UpdateUserActivity(req.Context, userID, activity)
	if err != nil {
		return a.Error(req, terrors.BadRequest("", "error decoding user", nil))
	}

	return req.Response(res)

}

func (a *API) DeleteUserActivity(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "user ID not supplied", nil))
	}

	aid, ok := a.Params(req)["activityID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "activity ID not supplied", nil))
	}

	userID := uuid.ID(id)
	activityID := uuid.ID(aid)

	if _, err := a.users.DeleteUserActivity(req.Context, uuid.ID(userID), activityID); err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.ResponseWithCode(nil, http.StatusNoContent)

}

func (a *API) GetUserActivities(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "user ID not supplied", nil))
	}

	userID := uuid.ID(id)

	as, err := a.users.ReadUserActivities(req.Context, uuid.ID(userID))
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.Response(as)

}
