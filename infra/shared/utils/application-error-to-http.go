package utils

import (
	"errors"
	errors2 "github.com/nahualventure/class-backend/core/app/shared/errors"
	userErrors "github.com/nahualventure/class-backend/core/app/user/domain/errors"
	"log"
	"net/http"
)

var ErrorCodeToHTTPStatus = map[errors2.ErrorCode]int{
	// Validation Errors
	errors2.ValidationError:             http.StatusBadRequest,
	errors2.DomainEntityValidationError: http.StatusBadRequest,

	// Authorization Errors
	errors2.Unauthorized: http.StatusUnauthorized,
	errors2.Forbidden:    http.StatusForbidden,

	// Infrastructure Errors
	errors2.InternalError: http.StatusInternalServerError,

	// User Errors
	userErrors.EmailAlreadyExistsError: http.StatusConflict,
	userErrors.UserNotFoundError:       http.StatusNotFound,
}

type HTTPErrorResponse struct {
	Error struct {
		Code      string                 `json:"code"`
		Message   string                 `json:"message"`
		Context   map[string]interface{} `json:"context,omitempty"`
		Timestamp string                 `json:"timestamp"`
	} `json:"error"`
	Status int `json:"-"`
}

func ApplicationErrorToHTTPResponse(err error) HTTPErrorResponse {
	var appErr errors2.ApplicationError
	if !errors.As(err, &appErr) {
		// Fallback for non-application errors
		return HTTPErrorResponse{
			Error: struct {
				Code      string                 `json:"code"`
				Message   string                 `json:"message"`
				Context   map[string]interface{} `json:"context,omitempty"`
				Timestamp string                 `json:"timestamp"`
			}{
				Code:      "INTERNAL_ERROR",
				Message:   "Internal server error",
				Timestamp: appErr.GetOccurredAt().Format("2006-01-02T15:04:05Z07:00"),
			},
			Status: http.StatusInternalServerError,
		}
	}

	// Log the full error with stack trace
	if appErr.Unwrap() != nil {
		log.Printf("Application Error: %+v", appErr.Unwrap())
	}

	// Convert string code back to ErrorCode type for map lookup
	errorCode := errors2.ErrorCode(appErr.GetCode())

	httpStatus, ok := ErrorCodeToHTTPStatus[errorCode]
	if !ok {
		httpStatus = http.StatusInternalServerError // default if mapping not found
	}

	message := appErr.GetMessage()
	// For internal errors, don't expose internal details
	if httpStatus == http.StatusInternalServerError {
		message = "Internal server error"
	}

	return HTTPErrorResponse{
		Error: struct {
			Code      string                 `json:"code"`
			Message   string                 `json:"message"`
			Context   map[string]interface{} `json:"context,omitempty"`
			Timestamp string                 `json:"timestamp"`
		}{
			Code:      appErr.GetCode(),
			Message:   message,
			Context:   appErr.GetContext(),
			Timestamp: appErr.GetOccurredAt().Format("2006-01-02T15:04:05Z07:00"),
		},
		Status: httpStatus,
	}
}
