package entities

import (
	appErrors "github.com/nahualventure/class-backend/core/app/shared/errors"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type User struct {
	ID        string    `validate:"required,uuid4"`
	Name      string    `validate:"required"`
	Email     string    `validate:"required,email"`
	CreatedAt time.Time `validate:"required"`
	UpdatedAt time.Time `validate:"required"`
}

func NewUser(id string, name string, email string, createdAt time.Time, updatedAt time.Time) (*User, error) {
	user := &User{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := validate.Struct(user); err != nil {
		var validationErrors validator.ValidationErrors
		errors.As(err, &validationErrors)
		errorMap := make(map[string]any)

		return nil, appErrors.NewDomainEntityValidationError("User domain model instance not valid", errorMap, err)
	}

	return user, nil
}
