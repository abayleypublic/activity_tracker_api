package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/errs"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/monzo/typhon"
)

func (a *API) GetChallenges(req typhon.Request) typhon.Response {

	cs := []challenges.Challenge{}
	if err := a.challenges.ReadAll(req.Context, &cs); err != nil {
		return errs.NotFoundResponse(req, err.Error())
	}

	return req.Response(cs)

}

func (a *API) GetChallenge(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "id not supplied")
	}

	challenge := challenges.Challenge{}
	if err := a.challenges.Read(req.Context, uuid.ID(id), &challenge); err != nil {
		return errs.NotFoundResponse(req, err.Error())
	}

	return req.Response(challenge)

}

func (a *API) PostChallenge(req typhon.Request) typhon.Response {

	var challenge challenges.Challenge
	if err := req.Decode(&challenge); err != nil {
		return errs.BadRequestResponse(req, "error decoding challenge")
	}
	challenge.ID = uuid.New()
	challenge.CreatedDate = time.Now().UTC()

	id, err := a.challenges.Create(req.Context, challenge)
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
	challenge := challenges.Challenge{}
	if err := a.challenges.Read(req.Context, challengeID, &challenge); err != nil {
		return errs.NotFoundResponse(req, err.Error())
	}

	// Stored challenge as slice of bytes
	subb, err := json.Marshal(challenge)
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
	c := challenges.Challenge{}
	if err = json.Unmarshal(modified, &challenge); err != nil {
		return errs.UnprocessableEntityResponse(req, "error unmarshalling challenge")
	}

	// Update user
	if err = a.challenges.Update(req.Context, c); err != nil {
		return errs.InternalServerResponse(req, err.Error())
	}

	return req.Response(c)

}

func (a *API) DeleteChallenge(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return errs.BadRequestResponse(req, "id not supplied")
	}

	if err := a.challenges.Delete(req.Context, uuid.ID(id)); err != nil {
		return errs.NotFoundResponse(req, err.Error())
	}

	return req.ResponseWithCode(nil, http.StatusNoContent)

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

	return req.ResponseWithCode(nil, http.StatusNoContent)
}
