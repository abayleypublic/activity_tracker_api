package users

import "errors"

var (
	ErrAlreadyExists = errors.New("user already exists")
	ErrNotFound      = errors.New("user not found")
	ErrUnknown       = errors.New("unknown error")
)
