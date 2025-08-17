package errors

import (
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
)

// ApplicationError interface - all errors implement this
type ApplicationError interface {
	error
	GetCode() string
	GetMessage() string
	GetContext() map[string]any
	GetOccurredAt() time.Time
	IsDomainError() bool
	Unwrap() error // For error unwrapping
}

// BaseError - common fields for all errors
type BaseError struct {
	Code       string         // Error code (e.g., "USER_NOT_FOUND")
	Message    string         // Human-readable message
	Context    map[string]any // Contextual data specific to this error
	OccurredAt time.Time      // When the error occurred
	Underlying error          // cockroachdb error with stack trace
}

func (e BaseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e BaseError) GetCode() string            { return e.Code }
func (e BaseError) GetMessage() string         { return e.Message }
func (e BaseError) GetContext() map[string]any { return e.Context }
func (e BaseError) GetOccurredAt() time.Time   { return e.OccurredAt }
func (e BaseError) Unwrap() error              { return e.Underlying }

// DetailedError returns error with full stack trace (for debugging)
func (e BaseError) DetailedError() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %s\n%+v", e.Code, e.Message, e.Underlying)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}


// BaseDomainError - business logic violations
type BaseDomainError struct {
	BaseError
}

func (e BaseDomainError) IsDomainError() bool { return true }

// InfrastructureError - technical failures
type InfrastructureError struct {
	BaseError
}

func (e InfrastructureError) IsDomainError() bool        { return false }
func (e InfrastructureError) GetContext() map[string]any { return nil }

func NewInfrastructureError(operation string, cause error) *InfrastructureError {
	underlying := cause
	if underlying == nil {
		underlying = errors.New(InternalError.String())
	} else {
		underlying = errors.Wrap(cause, InternalError.String())
	}

	return &InfrastructureError{
		BaseError: BaseError{
			Code:       InternalError.String(),
			Message:    operation,
			Context:    nil, // Infrastructure errors don't expose context
			OccurredAt: time.Now(),
			Underlying: underlying,
		},
	}
}

func NewValidationError(message string, errorMap map[string]any, cause error) *BaseDomainError {
	return &BaseDomainError{
		BaseError: BaseError{
			Code:       ValidationError.String(),
			Message:    message,
			Context:    errorMap,
			OccurredAt: time.Now(),
			Underlying: cause,
		},
	}
}

func NewDomainEntityValidationError(message string, errorMap map[string]any, cause error) *BaseDomainError {
	return &BaseDomainError{
		BaseError: BaseError{
			Code:       DomainEntityValidationError.String(),
			Message:    message,
			Context:    errorMap,
			OccurredAt: time.Now(),
			Underlying: cause,
		},
	}
}
