package api

const (
	NotFound       = "not found"
	NotAuthorised  = "not authorised"
	InternalServer = "internal server error"
)

type ErrorResponse struct {
	Cause string `json:"cause"`
}
