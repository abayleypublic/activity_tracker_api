package api

import (
	"net/http"
)

var (
	_ error = (*Error)(nil)
)

type Error struct {
	err   error  `json:"-"`
	Code  int    `json:"-"`
	Cause string `json:"message"`
}

func (e Error) Error() string {
	return e.err.Error()
}

func NewError(code int, cause string, err error) *Error {
	return &Error{
		err:   err,
		Code:  code,
		Cause: cause,
	}
}

func NotFound(cause string, err error) *Error {
	return NewError(http.StatusNotFound, cause, err)
}

func BadRequest(cause string, err error) *Error {
	return NewError(http.StatusBadRequest, cause, err)
}

func InternalServer(cause string, err error) *Error {
	return NewError(http.StatusInternalServerError, cause, err)
}

func Unauthorized(cause string, err error) *Error {
	return NewError(http.StatusUnauthorized, cause, err)
}

func Forbidden(cause string, err error) *Error {
	return NewError(http.StatusForbidden, cause, err)
}

func Conflict(cause string, err error) *Error {
	return NewError(http.StatusConflict, cause, err)
}

func UnprocessableEntity(cause string, err error) *Error {
	return NewError(http.StatusUnprocessableEntity, cause, err)
}
