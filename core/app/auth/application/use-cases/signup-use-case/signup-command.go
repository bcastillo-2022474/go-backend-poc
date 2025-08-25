package signup_use_case

import (
	"class-backend/core/app/shared/errors"
	"class-backend/core/app/shared/utils"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type CreateUserCommand struct {
	Name     string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=128"`
}

func NewCreateUserCommand(name string, email string, password string) (*CreateUserCommand, error) {
	command := &CreateUserCommand{
		Name:     name,
		Email:    email,
		Password: password,
	}

	if err := utils.ValidateStruct(validate, command); err != nil {
		return nil, errors.PropagateError(err)
	}

	return command, nil
}
