package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/monzo/typhon"
)

func (a *API) GetUsers(req typhon.Request) Response {

	users := []users.PartialUser{}
	if err := a.users.ReadAllRaw(req.Context, &users); err != nil {
		slog.ErrorContext(req.Context, "error reading users", "error", err)
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponse(users)
}

func (a *API) GetUser(req typhon.Request) Response {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	user, err := a.users.ReadUser(req.Context, service.ID(id))
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

	userID := service.ID(id)

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

	if err := a.users.Delete(req.Context, service.ID(id)); err != nil {
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

	if user.ID != service.ID(id) {
		return NewResponse(BadRequest("id in body does not match id in url", nil))
	}

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

func (a *API) JoinChallenge(req typhon.Request) Response {

	userID, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("user id not supplied", nil))
	}

	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("challenge id not supplied", nil))
	}

	_, err := a.users.JoinChallenge(req.Context, service.ID(userID), service.ID(id))
	if err != nil {
		switch err {
		case service.ErrResourceAlreadyExists:
			return NewResponse(Conflict(err.Error(), err))
		}
		return NewResponse(InternalServer(err.Error(), err))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)
}

func (a *API) LeaveChallenge(req typhon.Request) Response {

	userID, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("user id not supplied", nil))
	}

	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("challenge id not supplied", nil))
	}

	err := a.users.LeaveChallenge(req.Context, service.ID(userID), service.ID(id))
	if err != nil {
		switch err {
		case service.ErrResourceNotFound:
			return NewResponse(NotFound(err.Error(), err))
		}
		return NewResponse(InternalServer(err.Error(), err))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)
}

func (a *API) GetProfile(req typhon.Request) Response {

	u, err := service.GetActorContext(req.Context)
	if err != nil {
		return NewResponse(Unauthorized("could not determine user", err))
	}

	user := users.User{}
	if err = a.users.Read(req.Context, u.UserID, &user); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponse(user)

}
