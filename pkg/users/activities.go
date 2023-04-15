package users

import "github.com/monzo/typhon"

func (u *Users) GetUserActivities(req typhon.Request) typhon.Response {

	return req.Response("OK")

}
