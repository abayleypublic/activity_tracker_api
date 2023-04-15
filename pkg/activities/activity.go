package activities

import "github.com/monzo/typhon"

func (a *Activities) PutActivity(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (a *Activities) DeleteActivity(req typhon.Request) typhon.Response {

	return req.Response("OK")

}
