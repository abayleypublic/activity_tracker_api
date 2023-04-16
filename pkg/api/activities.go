package api

import "github.com/monzo/typhon"

func (a *API) GetActivities(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) PutActivity(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *API) DeleteActivity(req typhon.Request) typhon.Response {

	return req.Response("OK")

}
