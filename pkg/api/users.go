package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/AustinBayley/activity_tracker_api/pkg/users"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type PartialUser struct {
	ID        service.ID `json:"id" bson:"_id,omitempty"`
	FirstName string     `json:"firstName" bson:"firstName"`
	LastName  string     `json:"lastName" bson:"lastName"`
	Bio       string     `json:"bio" bson:"bio"`
}

func (a *API) GetUsers(req *gin.Context) {
	rawOpts := ListOptions{}
	if err := req.BindQuery(&rawOpts); err != nil {
		log.Error().
			Err(err).
			Msg("error binding query parameters")
	}

	opts := users.NewListOptions().
		SetLimit(rawOpts.Max).
		SetSkip(rawOpts.Page - 1)

	users := make([]PartialUser, 0, opts.Limit)
	if err := a.users.List(req, *opts, users); err != nil {
		log.Error().
			Err(err).
			Msg("error listing users")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	req.JSON(http.StatusOK, users)
}

func (a *API) GetUser(req *gin.Context) {
	id := req.Param("userID")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "user ID not supplied",
		})
		return
	}

	user := PartialUser{}
	if err := a.users.Get(req, service.ID(id), &user); err != nil {
		log.Error().
			Err(err).
			Str("userID", id).
			Msg("error getting user")

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

// PatchUser updates an existing user with the given ID using a JSON merge patch.
// The request body should contain a JSON merge patch that describes the changes to be made to the user.
// Returns a 400 Bad Request error if the ID is not supplied or if there is an error decoding the user.
// Returns a 404 Not Found error if the user with the given ID does not exist or if there is an error marshalling or unmarshalling the user.
// Returns a 204 No Content response if the user is successfully updated.
func (a *API) PatchUser(req *gin.Context) {
	id := req.Param("userID")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "user ID not supplied",
		})
		return
	}
	userID := service.ID(id)

	actor, ok := GetActorContext(req)
	if !ok {
		log.Error().
			Msg("failed to get actor from context")

		req.JSON(http.StatusUnauthorized, ErrorResponse{
			Cause: Unauthorised,
		})
		return
	}

	if userID != actor.UserID && !actor.Admin {
		log.Error().
			Str("ID", userID.ConvertID()).
			Msg("actor is not allowed to update user")

		req.JSON(http.StatusForbidden, ErrorResponse{
			Cause: "not allowed to update user",
		})
		return
	}

	// Read body as bytes
	bb, err := req.GetRawData()
	if err != nil {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "error reading request body",
		})
		return
	}

	// Get stored user
	user := users.Detail{}
	if err := a.users.Get(req, userID, &user); err != nil {
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

	// Marshal stored user to bytes
	subb, err := json.Marshal(user)
	if err != nil {
		req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Cause: "error marshalling user",
		})
		return
	}

	// Decode patch
	patch, err := jsonpatch.DecodePatch(bb)
	if err != nil {
		req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Cause: "could not decode patch",
		})
		return
	}

	// Apply patch
	modified, err := patch.Apply(subb)
	if err != nil {
		req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Cause: "could not apply patch",
		})
		return
	}

	// Unmarshal modified user
	if err := json.Unmarshal(modified, &user); err != nil {
		req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Cause: "error unmarshalling user",
		})
		return
	}

	// Update user
	if err := a.users.Update(req, &user); err != nil {
		log.Error().
			Err(err).
			Str("userID", string(user.ID)).
			Msg("error updating user")

		switch {
		case errors.Is(err, users.ErrNotFound):
			req.JSON(http.StatusNotFound, ErrorResponse{
				Cause: NotFound,
			})
			return
		case errors.Is(err, users.ErrValidation):
			req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Cause: Validation,
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

func (a *API) DeleteUser(req *gin.Context) {
	id := req.Param("userID")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "user ID not supplied",
		})
		return
	}
	userID := service.ID(id)

	actor, ok := GetActorContext(req)
	if !ok {
		log.Error().
			Msg("failed to get actor from context")

		req.JSON(http.StatusUnauthorized, ErrorResponse{
			Cause: Unauthorised,
		})
		return
	}

	if userID != actor.UserID && !actor.Admin {
		log.Error().
			Str("ID", userID.ConvertID()).
			Msg("actor is not allowed to delete user")

		req.JSON(http.StatusForbidden, ErrorResponse{
			Cause: "not allowed to delete user",
		})
		return
	}

	if err := a.users.Delete(req, userID); err != nil {
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

func (a *API) PostUser(req *gin.Context) {
	user := users.Detail{}
	if err := req.BindJSON(&user); err != nil {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "error binding user data",
		})
		return
	}
	user.ID = service.NewID()

	actor, ok := GetActorContext(req)
	if !ok {
		log.Error().
			Msg("error getting actor context")

		req.JSON(http.StatusUnauthorized, ErrorResponse{
			Cause: NotAuthorised,
		})
		return
	}

	if actor.UserID != "" && !actor.Admin {
		log.Error().
			Msg("user already exists")

		req.JSON(http.StatusConflict, ErrorResponse{
			Cause: Conflict,
		})

		return
	}

	if user.Email != actor.Email && !actor.Admin {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "email does not match authenticated user",
		})
		return
	}

	oID, err := a.users.Create(req, &user)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", user.Email).
			Msg("error creating user")

		switch {
		case errors.Is(err, users.ErrAlreadyExists):
			req.JSON(http.StatusConflict, ErrorResponse{
				Cause: "user already exists",
			})
			return
		case errors.Is(err, users.ErrValidation):
			req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Cause: Validation,
			})
			return
		}

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}
	user.ID = oID

	req.JSON(http.StatusCreated, user)
}

func (a *API) DownloadUserData(req *gin.Context) {
	id := req.Param("userID")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "user ID not supplied",
		})
		return
	}
	userID := service.ID(id)

	actor, ok := GetActorContext(req)
	if !ok {
		log.Error().
			Msg("failed to get actor from context")

		req.JSON(http.StatusUnauthorized, ErrorResponse{
			Cause: Unauthorised,
		})
		return
	}

	if userID != actor.UserID && !actor.Admin {
		log.Error().
			Str("ID", userID.ConvertID()).
			Msg("actor is not allowed to download user data")

		req.JSON(http.StatusForbidden, ErrorResponse{
			Cause: "not allowed to download user data",
		})
		return
	}

	// Get user details
	user := users.User{}
	if err := a.users.Get(req, userID, &user); err != nil {
		log.Error().
			Err(err).
			Str("userID", id).
			Msg("error getting user data")

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

	// Get user activities
	opts := activities.NewListOptions().SetUser(userID)
	activityList := make([]activities.Activity, 0)
	if err := a.activities.List(req, *opts, &activityList); err != nil {
		log.Error().
			Err(err).
			Str("userID", id).
			Msg("error getting user activities")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	// Get challenges created by user
	createdChallenges := make([]challenges.Detail, 0)
	if err := a.challenges.ListByCreator(req, userID, &createdChallenges); err != nil {
		log.Error().
			Err(err).
			Str("userID", id).
			Msg("error getting challenges created by user")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	// Package all user data
	userData := map[string]interface{}{
		"user":              user,
		"activities":        activityList,
		"createdChallenges": createdChallenges,
	}

	req.JSON(http.StatusOK, userData)
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

		actor, ok := GetActorContext(req)
		if !ok {
			log.Error().
				Msg("failed to get actor from context")

			req.JSON(http.StatusUnauthorized, ErrorResponse{
				Cause: Unauthorised,
			})
			return
		}

		if service.ID(userID) != actor.UserID && !actor.Admin {
			log.Error().
				Str("ID", userID).
				Msg("actor is not allowed to update user challenges")

			req.JSON(http.StatusForbidden, ErrorResponse{
				Cause: "not allowed to update user challenges",
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

		req.Status(http.StatusNoContent)
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
