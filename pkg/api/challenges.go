package api

import (
	"fmt"
	"net/http"

	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/AustinBayley/activity_tracker_api/pkg/uuid"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"
)

func (a *API) GetChallenges(req typhon.Request) typhon.Response {

	c, err := a.challenges.ReadChallenges(req.Context)
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.Response(c)

}

func (a *API) GetChallenge(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	c, err := a.challenges.ReadChallenge(req.Context, uuid.ID(id))
	if err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.Response(c)

}

func (a *API) PostChallenge(req typhon.Request) typhon.Response {

	var challenge challenges.Challenge
	if err := req.Decode(&challenge); err != nil {
		fmt.Println(err)
		return a.Error(req, terrors.BadRequest("", "error decoding challenge", nil))
	}

	id, err := a.challenges.CreateChallenge(req.Context, challenge)

	if err != nil {
		return a.Error(req, terrors.BadRequest("", err.Error(), nil))
	}

	return req.Response(id)

}

func (a *API) PatchChallenge(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) DeleteChallenge(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "id not supplied", nil))
	}

	if _, err := a.challenges.DeleteChallenge(req.Context, uuid.ID(id)); err != nil {
		return a.Error(req, terrors.NotFound("", err.Error(), nil))
	}

	return req.ResponseWithCode(nil, http.StatusOK)

}

func (a *API) PutMember(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "challenge ID not supplied", nil))
	}

	userID, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "user ID not supplied", nil))
	}

	res, err := a.challenges.AddMember(req.Context, uuid.ID(id), uuid.ID(userID))
	if err != nil {
		return a.Error(req, terrors.BadRequest("", err.Error(), nil))
	}
	if !res {
		return a.Error(req, terrors.BadRequest("", "error creating or updating member", nil))
	}

	return req.Response("Member created or updated successfully")
}

func (a *API) DeleteMember(req typhon.Request) typhon.Response {

	id, ok := a.Params(req)["id"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "challenge ID not supplied", nil))
	}

	userID, ok := a.Params(req)["userID"]
	if !ok {
		return a.Error(req, terrors.BadRequest("", "user ID not supplied", nil))
	}

	res, err := a.challenges.DeleteMember(req.Context, uuid.ID(id), uuid.ID(userID))
	if err != nil {
		return a.Error(req, terrors.BadRequest("", err.Error(), nil))
	}
	if !res {
		return a.Error(req, terrors.BadRequest("", "error deleting member", nil))
	}

	return req.ResponseWithCode(nil, http.StatusOK)
}
