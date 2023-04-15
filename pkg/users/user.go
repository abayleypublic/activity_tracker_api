package users

import "github.com/monzo/typhon"

func (u *Users) GetUser(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (u *Users) PatchUser(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (u *Users) DeleteUser(req typhon.Request) typhon.Response {

	return req.Response("OK")

}

func (u *Users) PutUser(req typhon.Request) typhon.Response {

	return req.Response("OK")

}
