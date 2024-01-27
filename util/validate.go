package validate

import "github.com/go-playground/validator/v10"

var validate *validator.Validate

func GetValidater() *validator.Validate {
	if validate == nil {
		validate = validator.New()
		validate.RegisterValidation("isempty", isEmpty)
	}
	return validate
}

func isEmpty(fl validator.FieldLevel) bool {
	return fl.Field().String() == ""
}
