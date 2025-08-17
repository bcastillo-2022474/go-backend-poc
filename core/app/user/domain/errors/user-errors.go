package errors

import (
	errors2 "class-backend/core/app/shared/errors"
	"time"

	"github.com/cockroachdb/errors"
)

const (
	UserNotFoundError       errors2.ErrorCode = "USER_NOT_FOUND"
	EmailAlreadyExistsError errors2.ErrorCode = "EMAIL_ALREADY_EXISTS"
)

func NewUserNotFoundError(userID string) *errors2.BaseDomainError {
	return &errors2.BaseDomainError{
		BaseError: errors2.BaseError{
			Code:    UserNotFoundError.String(),
			Message: "The requested user could not be found",
			Context: map[string]any{
				"user_id": userID,
			},
			OccurredAt: time.Now(),
			Underlying: errors.New(UserNotFoundError.String()), // Captures stack trace
		},
	}
}

func NewEmailAlreadyExistsError(email string) *errors2.BaseDomainError {
	return &errors2.BaseDomainError{
		BaseError: errors2.BaseError{
			Code:    EmailAlreadyExistsError.String(),
			Message: "A user with this email address already exists",
			Context: map[string]any{
				"email": email,
			},
			OccurredAt: time.Now(),
			Underlying: errors.New(EmailAlreadyExistsError.String()),
		},
	}
}
