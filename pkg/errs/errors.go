package errs

import (
	"net/http"

	"github.com/monzo/typhon"
)

type Error struct {
	Code  int    `json:"-"`
	Cause string `json:"cause"`
}

func Response(req typhon.Request, err Error) typhon.Response {
	return req.ResponseWithCode(err, err.Code)
}

func New(code int, cause string) Error {
	return Error{
		Code:  code,
		Cause: cause,
	}
}

func NotFound(cause string) Error {
	return New(http.StatusNotFound, cause)
}

func NotFoundResponse(req typhon.Request, cause string) typhon.Response {
	return Response(req, NotFound(cause))
}

func BadRequest(cause string) Error {
	return New(http.StatusBadRequest, cause)
}

func BadRequestResponse(req typhon.Request, cause string) typhon.Response {
	return Response(req, BadRequest(cause))
}

func InternalServer(cause string) Error {
	return New(http.StatusInternalServerError, cause)
}

func InternalServerResponse(req typhon.Request, cause string) typhon.Response {
	return Response(req, InternalServer(cause))
}

func Unauthorized(cause string) Error {
	return New(http.StatusUnauthorized, cause)
}

func UnauthorizedResponse(req typhon.Request, cause string) typhon.Response {
	return Response(req, Unauthorized(cause))
}

func Forbidden(cause string) Error {
	return New(http.StatusForbidden, cause)
}

func ForbiddenResponse(req typhon.Request, cause string) typhon.Response {
	return Response(req, Forbidden(cause))
}

func Conflict(cause string) Error {
	return New(http.StatusConflict, cause)
}

func ConflictResponse(req typhon.Request, cause string) typhon.Response {
	return Response(req, Conflict(cause))
}

func UnprocessableEntity(cause string) Error {
	return New(http.StatusUnprocessableEntity, cause)
}

func UnprocessableEntityResponse(req typhon.Request, cause string) typhon.Response {
	return Response(req, UnprocessableEntity(cause))
}
