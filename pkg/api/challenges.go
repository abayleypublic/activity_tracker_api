package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/monzo/typhon"
)

func (a *API) GetChallenges(req typhon.Request) Response {

	cs := []challenges.Challenge{}
	if err := a.challenges.ReadAll(req.Context, &cs); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponse(cs)

}

func (a *API) GetChallenge(req typhon.Request) Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	challenge := challenges.Challenge{}
	if err := a.challenges.Read(req.Context, service.ID(id), &challenge); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponse(challenge)

}

func (a *API) PostChallenge(req typhon.Request) Response {

	var challenge challenges.Challenge
	if err := req.Decode(&challenge); err != nil {
		return NewResponse(BadRequest("error decoding challenge", err))
	}
	challenge.ID = service.NewID()
	challenge.CreatedDate = time.Now().UTC()

	id, err := a.challenges.Create(req.Context, challenge)
	if err != nil {
		return NewResponse(InternalServer(err.Error(), err))
	}

	return NewResponse(id)

}

func (a *API) PatchChallenge(req typhon.Request) Response {

	// Get user ID
	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	challengeID := service.ID(id)

	// Get body & store as slice of bytes
	bb, err := req.BodyBytes(true)
	if err != nil {
		return NewResponse(BadRequest(err.Error(), err))
	}

	// Stored challenge
	challenge := challenges.Challenge{}
	if err := a.challenges.Read(req.Context, challengeID, &challenge); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	// Stored challenge as slice of bytes
	subb, err := json.Marshal(challenge)
	if err != nil {
		return NewResponse(UnprocessableEntity(err.Error(), err))
	}

	// Decode requested patch
	patch, err := jsonpatch.DecodePatch(bb)
	if err != nil {
		return NewResponse(UnprocessableEntity("could not decide request", err))
	}

	// Apply patch to stored challenge to get modified document
	modified, err := patch.Apply(subb)
	if err != nil {
		return NewResponse(UnprocessableEntity("could not apply patch", err))
	}

	// Unmarshal modified document into challenge struct
	c := challenges.Challenge{}
	if err = json.Unmarshal(modified, &challenge); err != nil {
		return NewResponse(UnprocessableEntity("error unmarshalling challenge", err))
	}

	// Update user
	if err = a.challenges.Update(req.Context, c); err != nil {
		return NewResponse(InternalServer(err.Error(), err))
	}

	return NewResponse(c)

}

func (a *API) DeleteChallenge(req typhon.Request) Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("id not supplied", nil))
	}

	if err := a.challenges.Delete(req.Context, service.ID(id)); err != nil {
		return NewResponse(NotFound(err.Error(), err))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)

}

func (a *API) PutMember(req typhon.Request) Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("challenge id not supplied", nil))
	}

	userID, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("user id not supplied", nil))
	}

	d, err := a.challenges.AppendAttribute(req.Context, service.ID(id), "members", service.ID(userID))
	if err != nil {
		switch err {
		case service.ErrResourceAlreadyExists:
			return NewResponse(Conflict(err.Error(), err))
		case service.ErrResourceNotFound:
			return NewResponse(NotFound(err.Error(), err))
		}
		return NewResponse(InternalServer(err.Error(), err))
	}

	log.Println(d)

	return NewResponseWithCode(nil, http.StatusNoContent)
}

func (a *API) DeleteMember(req typhon.Request) Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return NewResponse(BadRequest("challenge id not supplied", nil))
	}

	userID, ok := a.Params(req)["userID"]
	if !ok {
		return NewResponse(BadRequest("user id not supplied", nil))
	}

	err := a.challenges.DeleteMember(req.Context, service.ID(id), service.ID(userID))
	if err != nil {
		switch err {
		case service.ErrResourceNotFound:
			return NewResponse(NotFound(err.Error(), err))
		}
		return NewResponse(InternalServer(err.Error(), err))
	}

	return NewResponseWithCode(nil, http.StatusNoContent)
}
