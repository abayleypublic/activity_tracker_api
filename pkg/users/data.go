package users

import "github.com/monzo/typhon"

func (u *Users) DownloadUserData(req typhon.Request) typhon.Response {

	return req.Response("OK")

}
