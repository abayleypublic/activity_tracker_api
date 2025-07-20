package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/AustinBayley/activity_tracker_api/pkg/activities"
	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (a *API) GetChallenges(req *gin.Context) {
	rawOpts := ListOptions{}
	if err := req.BindQuery(&rawOpts); err != nil {
		log.Error().
			Err(err).
			Msg("error binding query parameters")
	}

	opts := challenges.NewListOptions().
		SetLimit(rawOpts.Max).
		SetSkip(rawOpts.Page - 1)

	cs := []challenges.Challenge{}
	if err := a.challenges.List(req, *opts, &cs); err != nil {
		log.Error().
			Err(err).
			Msg("error listing challenges")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	req.JSON(http.StatusOK, cs)
}

func (a *API) GetChallenge(req *gin.Context) {
	id := req.Param("id")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "challenge ID not supplied",
		})
		return
	}

	challenge := challenges.Challenge{}
	if err := a.challenges.Get(req, service.ID(id), &challenge); err != nil {
		log.Error().
			Err(err).
			Str("challengeID", id).
			Msg("error getting challenge")

		if errors.Is(err, challenges.ErrNotFound) {
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

	req.JSON(http.StatusOK, challenge)
}

func (a *API) PostChallenge(req *gin.Context) {
	var challenge challenges.Challenge
	if err := req.BindJSON(&challenge); err != nil {
		log.Error().
			Err(err).
			Msg("error binding JSON to challenge")

		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "invalid request body",
		})
		return
	}
	challenge.ID = service.NewID()

	if err := a.challenges.Create(req, &challenge); err != nil {
		log.Error().
			Err(err).
			Msg("error creating challenge")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	req.JSON(http.StatusCreated, challenge)
}

func (a *API) PatchChallenge(req *gin.Context) {
	id := req.Param("id")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "challenge ID not supplied",
		})
		return
	}

	challengeID := service.ID(id)

	// Read body as bytes
	bb, err := req.GetRawData()
	if err != nil {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "error reading request body",
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

	operations := []challenges.Operation{}

	// Iterates to find any operations that set the members
	for i, op := range patch {
		path, err := op.Path()
		if err != nil {
			req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Cause: "could not get path from operation",
			})
			return
		}

		if strings.HasPrefix(path, "/members") {
			value, err := op.ValueInterface()
			if err != nil || value == nil {
				req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
					Cause: "could not get value from operation",
				})
				return
			}

			id, ok := value.(string)
			if !ok {
				req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
					Cause: "value is not a string",
				})
				return
			}

			operations = append(operations, challenges.SetMemberOperation{
				Challenge: challengeID,
				User:      service.ID(id),
				Member:    op.Kind() == "add",
			})

			patch = append(patch[:i], patch[i+1:]...)
		}
	}

	if len(patch) > 0 {
		// Get stored challenge
		challenge := challenges.Detail{}
		if err := a.challenges.Get(req, challengeID, &challenge); err != nil {
			if errors.Is(err, challenges.ErrNotFound) {
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

		// Marshal stored challenge to bytes
		subb, err := json.Marshal(challenge)
		if err != nil {
			req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Cause: "error marshalling challenge",
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

		// Unmarshal modified challenge
		if err := json.Unmarshal(modified, &challenge); err != nil {
			req.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Cause: "error unmarshalling challenge",
			})
			return
		}

		operations = append(operations, challenges.SetDetailOperation{
			Detail: challenge,
		})
	}

	// Update challenge
	if err := a.challenges.Update(req, operations...); err != nil {
		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	req.JSON(http.StatusNoContent, nil)
}

func (a *API) DeleteChallenge(req *gin.Context) {
	id := req.Param("id")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "challenge ID not supplied",
		})
		return
	}

	if err := a.challenges.Delete(req, service.ID(id)); err != nil {
		log.Error().
			Err(err).
			Msg("failed to delete challenge")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	req.JSON(http.StatusNoContent, nil)
}

func (a *API) GetProgress(req *gin.Context) {
	id := req.Param("id")
	if id == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "challenge ID not supplied",
		})
		return
	}

	uID := req.Param("userID")
	if uID == "" {
		req.JSON(http.StatusBadRequest, ErrorResponse{
			Cause: "user ID not supplied",
		})
		return
	}

	userID := service.ID(uID)

	challenge := challenges.Challenge{}
	if err := a.challenges.Get(req, service.ID(id), &challenge); err != nil {
		log.Error().
			Err(err).
			Str("challengeID", id).
			Msg("error getting challenge")

		if errors.Is(err, challenges.ErrNotFound) {
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

	if !slices.Contains(challenge.Members, userID) {
		req.JSON(http.StatusNotFound, ErrorResponse{
			Cause: NotFound,
		})
		return
	}

	opts := activities.ListOptions{
		User: &userID,
	}

	acts := []activities.Activity{}
	if err := a.activities.List(req, opts, &acts); err != nil {
		log.Error().
			Err(err).
			Str("userID", string(userID)).
			Str("challengeID", id).
			Msg("error listing user activities")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	progress, err := challenge.Target.Evaluate(req, acts)
	if err != nil {
		log.Error().
			Err(err).
			Str("userID", string(userID)).
			Str("challengeID", id).
			Msg("error evaluating challenge progress")

		req.JSON(http.StatusInternalServerError, ErrorResponse{
			Cause: InternalServer,
		})
		return
	}

	req.JSON(http.StatusOK, progress)
}
