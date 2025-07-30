package validate

import "github.com/go-playground/validator/v10"

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func Struct(s interface{}) error {
	return Validate.Struct(s)
}
