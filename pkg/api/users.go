package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/monzo/typhon"
)

func (a *API) GetUsers(req typhon.Request) Response {

	users := []users.User{}
	if err := a.users.ReadAll(req.Context, &users); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponse(users)
}

func (a *API) GetUser(req typhon.Request) Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	user := users.User{}
	err := a.users.Read(req.Context, uuid.ID(id), &user)
	if err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponse(user)

}

// PatchUser updates an existing user with the given ID using a JSON merge patch.
// The request body should contain a JSON merge patch that describes the changes to be made to the user.
// Returns a 400 Bad Request error if the ID is not supplied or if there is an error decoding the user.
// Returns a 404 Not Found error if the user with the given ID does not exist or if there is an error marshalling or unmarshalling the user.
// Returns a 204 No Content response if the user is successfully updated.
func (a *API) PatchUser(req typhon.Request) Response {

	// Get user ID
	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	userID := uuid.ID(id)

	// Get body & store as slice of bytes
	bb, err := req.BodyBytes(true)
	if err != nil {
		return NewResponse(BadRequest(err.Error(), err))
	}

	// Stored user
	user := users.User{}
	err = a.users.Read(req.Context, userID, &user)
	if err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	// Stored user as slice of bytes
	subb, err := json.Marshal(user)
	if err != nil {
		return NewResponse(UnprocessableEntity(err.Error(), err))
	}

	// Decode requested patch
	patch, err := jsonpatch.DecodePatch(bb)
	if err != nil {
		return NewResponse(UnprocessableEntity("could not decode request", err))
	}

	// Apply patch to stored user to get modified document
	modified, err := patch.Apply(subb)
	if err != nil {
		return NewResponse(UnprocessableEntity("could not apply patch", err))
	}

	// Unmarshal modified document into user struct
	user = users.User{}
	if err = json.Unmarshal(modified, &user); err != nil {
		return NewResponse(UnprocessableEntity("error unmarshalling user", err))
	}

	// Update user
	if err = a.users.Update(req.Context, user); err != nil {
		return NewResponse(InternalServer(err.Error(), err))
	}

	return NewResponse(user)

}

func (a *API) DeleteUser(req typhon.Request) Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	if err := a.users.Delete(req.Context, uuid.ID(id)); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)

}

func (a *API) PutUser(req typhon.Request) Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	user := users.User{}
	if err := req.Decode(&user); err != nil {
		return NewResponse(UnprocessableEntity("error decoding user", err))
	}

	if user.ID != uuid.ID(id) {
		return NewResponse(BadRequest("id in body does not match id in url", nil))
	}

	user.CreatedDate = time.Now().UTC()

	res, err := a.users.Create(req.Context, user)
	if err != nil {
		switch err {
		case service.ErrResourceAlreadyExists:
			return NewResponse(Conflict("user already exists", err))
		default:
			return NewResponse(InternalServer("error creating user", err))
		}
	}

	return NewResponseWithCode(res, http.StatusOK)

}

func (a *API) DownloadUserData(req typhon.Request) Response {

	return NewResponse("OK")

}

func (a *API) GetUserActivity(req typhon.Request) Response {
	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	aid, ok := a.Params(req)["activityID"]
	if !ok {
		return NewResponse(BadRequest("activity id not supplied", nil))
	}

	userID := uuid.ID(id)
	activityID := uuid.ID(aid)

	res, err := a.users.ReadUserActivity(req.Context, userID, activityID)
	if err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponse(res)
}

func (a *API) PostUserActivity(req typhon.Request) Response {
	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}
	userID := uuid.ID(id)

	activity := activities.Activity{}
	if err := req.Decode(&activity); err != nil {
		return NewResponse(UnprocessableEntity("error decoding activity", err))
	}
	activity.ID = uuid.New()

	res, err := a.users.CreateUserActivity(req.Context, userID, activity)
	if err != nil {
		return NewResponse(InternalServer(err.Error(), err))
	}

	return NewResponse(res)
}

func (a *API) PatchUserActivity(req typhon.Request) Response {
	// Get user ID
	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	// Get activity ID
	aid, ok := a.Params(req)["activityID"]
	if !ok {
		return NewResponse(BadRequest("activity id not supplied", nil))
	}

	userID := uuid.ID(id)
	activityID := uuid.ID(aid)

	// Get body & store as slice of bytes
	bb, err := req.BodyBytes(true)
	if err != nil {
		return NewResponse(BadRequest(err.Error(), err))
	}

	// Stored activity
	sa, err := a.users.ReadUserActivity(req.Context, userID, activityID)
	if err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	// Stored activity as slice of bytes
	sabb, err := json.Marshal(sa)
	if err != nil {
		return NewResponse(UnprocessableEntity("error marshalling stored activity", err))
	}

	// Decode requested patch
	patch, err := jsonpatch.DecodePatch(bb)
	if err != nil {
		return NewResponse(UnprocessableEntity("could not decode request", err))
	}

	// Apply patch to stored activity to get modified document
	modified, err := patch.Apply(sabb)
	if err != nil {
		return NewResponse(UnprocessableEntity("could not apply patch", err))
	}

	// Unmarshal modified document into user struct
	var activity activities.Activity
	if err = json.Unmarshal(modified, &activity); err != nil {
		return NewResponse(UnprocessableEntity("error unmarshalling activity", err))
	}

	// Update activity
	res, err := a.users.UpdateUserActivity(req.Context, userID, activity)
	if err != nil {
		return NewResponse(InternalServer(err.Error(), err))
	}

	return NewResponse(res)
}

func (a *API) DeleteUserActivity(req typhon.Request) Response {
	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("user id not supplied", nil))
	}

	aid, ok := a.Params(req)["activityID"]
	if !ok {
		return NewResponse(BadRequest("activity id not supplied", nil))
	}

	if err := a.users.DeleteUserActivity(req.Context, uuid.ID(id), uuid.ID(aid)); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)
}

func (a *API) GetUserActivities(req typhon.Request) Response {
	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("user id not supplied", nil))
	}

	as, err := a.users.ReadUserActivities(req.Context, uuid.ID(id))
	if err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponse(as)
}
