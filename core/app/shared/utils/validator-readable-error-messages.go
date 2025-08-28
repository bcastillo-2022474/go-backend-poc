package utils

import (
	"errors"
	appErrors "github.com/nahualventure/class-backend/core/app/shared/errors"

	"github.com/go-playground/validator/v10"
)

func MsgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email address"
	case "min":
		return "Too short (minimum " + fe.Param() + " characters)"
	case "max":
		return "Too long (maximum " + fe.Param() + " characters)"
	default:
		return fe.Error() // fallback to default error
	}
}

func ValidateStruct(validate *validator.Validate, command interface{}) *appErrors.BaseDomainError {
	err := validate.Struct(command)

	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		errorMap := make(map[string]any)

		for _, fe := range validationErrors {
			errorMap[fe.Field()] = MsgForTag(fe)
		}

		return appErrors.NewValidationError("Invalid user creation request", errorMap, err)
	}

	return appErrors.NewValidationError("Invalid user creation request", nil, err)
}
