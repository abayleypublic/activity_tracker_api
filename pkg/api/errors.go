package api

const (
	NotFound       = "not found"
	NotAuthorised  = "not authorised"
	InternalServer = "internal server error"
	Invalid        = "invalid request"
	Conflict       = "conflict"
	Forbidden      = "forbidden"
	Unauthorised   = "unauthorised"
	Validation     = "validation error"
)

type ErrorResponse struct {
	Cause string `json:"cause"`
}
