package challenges

import (
	"errors"
)

var (
	ErrAlreadyExists = errors.New("challenge already exists")
	ErrNotFound      = errors.New("challenge not found")
	ErrUnknown       = errors.New("unknown error")
	ErrInvalid       = errors.New("invalid")
)
