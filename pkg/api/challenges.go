package api

import (
	"encoding/json"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/errs"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/monzo/typhon"
)

func (a *API) GetChallenges(req typhon.Request) typhon.Response {

	c, err := a.challenges.ReadChallenges(req.Context)
	if err != nil {
		return errs.NotFoundResponse(req, err.Error())
	}

	return req.Response(c)

}

func (a *API) GetChallenge(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "id not supplied")
	}

	c, err := a.challenges.ReadChallenge(req.Context, uuid.ID(id))
	if err != nil {
		return errs.NotFoundResponse(req, err.Error())
	}

	return req.Response(c)

}

func (a *API) PostChallenge(req typhon.Request) typhon.Response {

	var challenge challenges.Challenge
	if err := req.Decode(&challenge); err != nil {
		return errs.BadRequestResponse(req, "error decoding challenge")
	}

	id, err := a.challenges.CreateChallenge(req.Context, challenge)
	if err != nil {
		return errs.InternalServerResponse(req, err.Error())
	}

	return req.Response(id)

}

func (a *API) PatchChallenge(req typhon.Request) typhon.Response {

	// Get user ID
	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "id not supplied")
	}

	challengeID := uuid.ID(id)

	// Get body & store as slice of bytes
	bb, err := req.BodyBytes(true)
	if err != nil {
		return errs.BadRequestResponse(req, err.Error())
	}

	// Stored challenge
	su, err := a.challenges.ReadChallenge(req.Context, challengeID)
	if err != nil {
		return errs.NotFoundResponse(req, err.Error())
	}

	// Stored challenge as slice of bytes
	subb, err := json.Marshal(su)
	if err != nil {
		return errs.UnprocessableEntityResponse(req, err.Error())
	}

	// Decode requested patch
	patch, err := jsonpatch.DecodePatch(bb)
	if err != nil {
		return errs.UnprocessableEntityResponse(req, "could not decide request")
	}

	// Apply patch to stored challenge to get modified document
	modified, err := patch.Apply(subb)
	if err != nil {
		return errs.UnprocessableEntityResponse(req, "could not apply patch")
	}

	// Unmarshal modified document into challenge struct
	challenge := challenges.Challenge{}
	if err = json.Unmarshal(modified, &challenge); err != nil {
		return errs.UnprocessableEntityResponse(req, "error unmarshalling challenge")
	}

	// Update user
	if err = a.challenges.UpdateChallenge(req.Context, challenge); err != nil {
		return errs.InternalServerResponse(req, err.Error())
	}

	return req.Response(challenge)

}

func (a *API) DeleteChallenge(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "id not supplied")
	}

	if err := a.challenges.DeleteChallenge(req.Context, uuid.ID(id)); err != nil {
		return errs.NotFoundResponse(req, err.Error())
	}

	return req.ResponseWithCode(nil, http.StatusOK)

}

func (a *API) PutMember(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "challenge id not supplied")
	}

	userID, ok := a.Params(req)["userID"]
	if !ok {
		return errs.BadRequestResponse(req, "user id not supplied")
	}

	err := a.challenges.AddMember(req.Context, uuid.ID(id), uuid.ID(userID))
	if err != nil {
		return errs.InternalServerResponse(req, err.Error())
	}

	return req.Response("Member created or updated successfully")
}

func (a *API) DeleteMember(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "challenge id not supplied")
	}

	userID, ok := a.Params(req)["userID"]
	if !ok {
		return errs.BadRequestResponse(req, "user id not supplied")
	}

	err := a.challenges.DeleteMember(req.Context, uuid.ID(id), uuid.ID(userID))
	if err != nil {
		return errs.InternalServerResponse(req, err.Error())
	}

	return req.ResponseWithCode(nil, http.StatusOK)
}
