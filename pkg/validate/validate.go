package validate

import "github.com/go-playground/validator/v10"

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

// Struct validates the given struct using the global validator instance.
func Struct(s interface{}) error {
	return Validate.Struct(s)
}
