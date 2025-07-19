package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (a *API) GetUsers(req *gin.Context) {

	users := []users.PartialUser{}
	if err := a.users.ReadAllRaw(req.Context, &users); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponse(users)
}

func (a *API) GetUser(req *gin.Context) {

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
func (a *API) PatchUser(req *gin.Context) {

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

func (a *API) DeleteUser(req *gin.Context) {
	id := req.Param("userID")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "user ID not supplied",
		})
		return
	}

	if err := a.users.Delete(req, service.ID(id)); err != nil {
		log.Error().
			Err(err).
			Str("userID", id).
			Msg("error deleting user")

		if errors.Is(err, users.ErrNotFound) {
			req.JSON(http.StatusNotFound, ErrorResponse{
				Cause: NotFound,
			})
			return
		}

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	req.JSON(http.StatusNoContent, nil)
}

func (a *API) PutUser(req *gin.Context) {

	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	user := users.User{}
	if err := req.BindJSON(&user); err != nil {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "error binding user data",
		})
		return
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

func (a *API) DownloadUserData(req *gin.Context) {
	req.JSON(http.StatusNotImplemented, ErrorResponse{
		Cause: "not implemented",
	})
}

func (a *API) SetChallengeMembership(member bool) gin.HandlerFunc {
	return func(req *gin.Context) {
		userID := req.Param("userID")
		if userID == "" {
			req.JSON(http.StatusBadRequest, ErrorResponse{
				Cause: "user ID not supplied",
			})
			return
		}

		challengeID := req.Param("id")
		if challengeID == "" {
			req.JSON(http.StatusBadRequest, ErrorResponse{
				Cause: "challenge ID not supplied",
			})
			return
		}

		op := challenges.SetMemberOperation{
			User:      service.ID(userID),
			Challenge: service.ID(challengeID),
			Member:    member,
		}

		if err := a.challenges.Update(req, op); err != nil {
			log.Error().
				Err(err).
				Str("userID", userID).
				Str("challengeID", challengeID).
				Bool("member", member).
				Msg("error setting challenge membership")

			switch err {
			case challenges.ErrNotFound:
				req.JSON(http.StatusNotFound, ErrorResponse{
					Cause: NotFound,
				})
				return
			}

			req.JSON(http.StatusInternalServerError, ErrorResponse{
				Cause: InternalServer,
			})
			return
		}
	}
}

func (a *API) GetProfile(req *gin.Context) {
	ctx, ok := GetActorContext(req)
	if !ok {
		req.JSON(http.StatusUnauthorized, ErrorResponse{
			Cause: NotAuthorised,
		})
		return
	}

	user := users.User{}
	if err := a.users.Get(req, ctx.UserID, &user); err != nil {
		log.Error().
			Err(err).
			Str("userID", string(ctx.UserID)).
			Msg("error getting user profile")

		if errors.Is(err, users.ErrNotFound) {
			req.JSON(http.StatusNotFound, ErrorResponse{
				Cause: NotFound,
			})
			return
		}

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	req.JSON(http.StatusOK, user)
}
