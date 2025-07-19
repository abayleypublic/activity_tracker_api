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

	id, ok := req.Params["userID"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}
	userID := service.ID(id)

	activity := activities.Activity{}
	if err := req.Decode(&activity); err != nil {
		return NewResponse(UnprocessableEntity("error decoding activity", err))
	}
	activity.ID = service.NewID()
	activity.UserID = userID

	res, err := a.users.CreateUserActivity(req.Context, activity)
	if err != nil {
		return NewResponse(InternalServer(err.Error(), err))
	}

	return NewResponseWithCode(res, http.StatusCreated)
}

func (a *API) PatchUserActivity(req *gin.Context) {
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

	userID := service.ID(id)
	activityID := service.ID(aid)

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

func (a *API) DeleteUserActivity(req *gin.Context) {
	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("user id not supplied", nil))
	}

	aid, ok := a.Params(req)["activityID"]
	if !ok {
		return NewResponse(BadRequest("activity id not supplied", nil))
	}

	if err := a.users.DeleteUserActivity(req.Context, service.ID(id), service.ID(aid)); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)
}

func (a *API) GetUserActivities(req *gin.Context) {
	id, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("user id not supplied", nil))
	}

	as, err := a.users.ReadUserActivities(req.Context, service.ID(id))
	if err != nil {
		return NewResponse(err)
	}

	return NewResponse(as)
}
