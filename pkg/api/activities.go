package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (a *API) GetActivity(req *gin.Context) {
	aid := req.Param("activityID")
	if aid == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "activity id not supplied",
		})
		return
	}
	activityID := service.ID(aid)

	activity := activities.Activity{}
	if err := a.activities.Get(req, activityID, &activity); err != nil {
		log.Error().
			Err(err).
			Str("activityID", aid).
			Msg("error getting activity")

		if errors.Is(err, activities.ErrNotFound) {
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

	req.JSON(http.StatusOK, activity)
}

func (a *API) PostUserActivity(req *gin.Context) {
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

	if actor.UserID != userID && !actor.Admin {
		log.Error().
			Str("userID", userID.ConvertID()).
			Msg("actor is not allowed to create activity for user")

		req.JSON(http.StatusForbidden, ErrorResponse{
			Cause: "not allowed to create activity for user",
		})
		return
	}

	activity := activities.Activity{}
	if err := req.BindJSON(&activity); err != nil {
		log.Error().
			Err(err).
			Msg("error binding request body")

		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "invalid request body",
		})
		return
	}
	activity.ID = service.NewID()
	activity.UserID = userID

	oid, err := a.activities.Create(req, &activity)
	if err != nil {
		log.Error().
			Err(err).
			Str("userID", string(userID)).
			Msg("error creating activity")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}
	activity.ID = oid

	req.JSON(http.StatusCreated, activity)
}

func (a *API) PatchActivity(req *gin.Context) {
	id := req.Param("activityID")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "activity ID not supplied",
		})
		return
	}
	aID := service.ID(id)

	stored := activities.Activity{}
	if err := a.activities.Get(req, aID, &stored); err != nil {
		log.Error().
			Err(err).
			Str("activityID", id).
			Msg("error getting activity")

		if errors.Is(err, activities.ErrNotFound) {
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

	actor, ok := GetActorContext(req)
	if !ok {
		log.Error().
			Str("ID", stored.ID.ConvertID()).
			Msg("failed to get actor from context")

		req.JSON(http.StatusUnauthorized, ErrorResponse{
			Cause: Unauthorised,
		})
		return
	}

	if stored.UserID != actor.UserID && !actor.Admin {
		log.Error().
			Str("ID", stored.ID.ConvertID()).
			Msg("actor is not allowed to update activity")

		req.JSON(http.StatusForbidden, ErrorResponse{
			Cause: "not allowed to update activity",
		})
		return
	}

	// Stored activity as slice of bytes
	sabb, err := json.Marshal(stored)
	if err != nil {
		log.Error().
			Err(err).
			Msg("error marshalling activity")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	// Get body & store as slice of bytes
	bb, err := req.GetRawData()
	if err != nil {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "error reading request body",
		})
		return
	}

	// Decode requested patch
	patch, err := jsonpatch.DecodePatch(bb)
	if err != nil {
		log.Error().
			Err(err).
			Msg("error decoding patch")

		req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Cause: "could not decode patch",
		})
		return
	}

	// Apply patch to stored activity to get modified document
	modified, err := patch.Apply(sabb)
	if err != nil {
		log.Error().
			Err(err).
			Msg("error applying patch")

		req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Cause: "could not apply patch",
		})
		return
	}

	// Unmarshal modified document into user struct
	var activity activities.Activity
	if err = json.Unmarshal(modified, &activity); err != nil {
		log.Error().
			Err(err).
			Msg("error unmarshalling activity")

		req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Cause: "error unmarshalling activity",
		})
		return
	}

	// Update activity
	if err = a.activities.Update(req, activity); err != nil {
		log.Error().
			Err(err).
			Str("activityID", id).
			Msg("error updating activity")

		if errors.Is(err, activities.ErrNotFound) {
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

func (a *API) DeleteActivity(req *gin.Context) {
	id := req.Param("activityID")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "activity ID not supplied",
		})
		return
	}
	aID := service.ID(id)

	activity := activities.Activity{}
	if err := a.activities.Get(req, service.ID(id), &activity); err != nil {
		log.Error().
			Err(err).
			Str("activityID", id).
			Msg("error getting activity")

		if errors.Is(err, activities.ErrNotFound) {
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

	actor, ok := GetActorContext(req)
	if !ok {
		log.Error().
			Str("ID", id).
			Msg("failed to get actor from context")

		req.JSON(http.StatusUnauthorized, ErrorResponse{
			Cause: Unauthorised,
		})
		return
	}

	if activity.UserID != actor.UserID && !actor.Admin {
		log.Error().
			Str("ID", activity.ID.ConvertID()).
			Msg("actor is not allowed to delete activity")

		req.JSON(http.StatusForbidden, ErrorResponse{
			Cause: "not allowed to delete activity",
		})
		return
	}

	opts := activities.ActivityDeleteOpts{
		ID: &aID,
	}

	if err := a.activities.Delete(req, opts); err != nil {
		log.Error().
			Err(err).
			Str("activityID", id).
			Msg("error deleting activity")

		if errors.Is(err, activities.ErrNotFound) {
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

func (a *API) GetUserActivities(req *gin.Context) {
	id := req.Param("userID")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "user ID not supplied",
		})
		return
	}

	rawOpts := ListOptions{}
	if err := req.BindQuery(&rawOpts); err != nil {
		log.Error().
			Err(err).
			Msg("error binding query parameters")
	}

	opts := activities.NewListOptions().
		SetLimit(rawOpts.Max).
		SetSkip(rawOpts.Page - 1).
		SetUser(service.ID(id))

	activities := make([]activities.Activity, 0, opts.Limit)
	if err := a.activities.List(req, *opts, &activities); err != nil {
		log.Error().
			Err(err).
			Msg("error listing user activities")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	req.JSON(http.StatusOK, activities)
}
