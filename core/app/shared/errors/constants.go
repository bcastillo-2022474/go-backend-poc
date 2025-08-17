package errors

type ErrorCode string

const (
	// Validation Errors
	ValidationError             ErrorCode = "VALIDATION_ERROR"
	DomainEntityValidationError ErrorCode = "DOMAIN_ENTITY_VALIDATION_ERROR"

	// Authorization Errors
	Unauthorized ErrorCode = "UNAUTHORIZED"
	Forbidden    ErrorCode = "FORBIDDEN"

	// Infrastructure Errors
	InternalError ErrorCode = "INTERNAL_ERROR"
)

// String returns the string representation of the error code
func (e ErrorCode) String() string {
	return string(e)
}
