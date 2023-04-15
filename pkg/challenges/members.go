package challenges

import "github.com/monzo/typhon"

func (c *Challenges) GetMembers(req typhon.Request) typhon.Response {

	return req.Response("OK")

}
