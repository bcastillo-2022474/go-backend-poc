package utils

import (
	errors2 "class-backend/core/app/shared/errors"
	userErrors "class-backend/core/app/user/domain/errors"
	commonv1 "class-backend/proto/generated/go/common/v1"
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrorCodeToGrpcCode = map[errors2.ErrorCode]codes.Code{
	// Validation Errors
	errors2.ValidationError:             codes.InvalidArgument,
	errors2.DomainEntityValidationError: codes.InvalidArgument,

	// Authorization Errors
	errors2.Unauthorized: codes.Unauthenticated,
	errors2.Forbidden:    codes.PermissionDenied,

	// Infrastructure Errors
	errors2.InternalError: codes.Internal,

	// User Errors
	userErrors.EmailAlreadyExistsError: codes.AlreadyExists,
	userErrors.UserNotFoundError:       codes.NotFound,
}

// ApplicationErrorToProtoDetails converts an ApplicationError to a gRPC ErrorDetail proto message
// This function belongs in the infrastructure layer to maintain clean architecture boundaries
func ApplicationErrorToProtoDetails(appErr errors2.ApplicationError) *commonv1.ErrorDetail {
	// Convert dynamic context to protobuf struct
	var contextStruct *structpb.Struct

	// Only process context if it exists and has content
	if appErr.GetContext() != nil && len(appErr.GetContext()) > 0 {
		var err error
		contextStruct, err = structpb.NewStruct(appErr.GetContext())
		if err != nil {
			// If conversion fails, create a fallback context
			contextStruct, _ = structpb.NewStruct(map[string]interface{}{
				"context_error": "Failed to serialize original context",
			})
		}
	}
	// If context is nil or empty, contextStruct remains nil

	return &commonv1.ErrorDetail{
		Code:      appErr.GetCode(),
		Message:   appErr.GetMessage(),
		Context:   contextStruct, // nil for infra, populated for domain
		Timestamp: timestamppb.New(appErr.GetOccurredAt()),
	}
}

func ApplicationErrorToGrpcStatus(err error) error {
	var appErr errors2.ApplicationError
	if !errors.As(err, &appErr) {
		// Fallback for non-application errors
		return status.Errorf(codes.Internal, "Internal server error")
	}

	// Log the full error with stack trace before converting to gRPC
	if appErr.Unwrap() != nil {
		log.Printf("Application Error: %+v", appErr.Unwrap()) // %+v gives full stack trace with cockroach/errors
	}
	// Convert string code back to ErrorCode type for map lookup
	errorCode := errors2.ErrorCode(appErr.GetCode())

	grpcCode, ok := ErrorCodeToGrpcCode[errorCode]
	if !ok {
		grpcCode = codes.Internal // default if mapping not found
	}

	// For internal errors, don't expose internal details
	if grpcCode == codes.Internal {
		return status.Error(grpcCode, "Internal server error")
	}

	st := status.New(grpcCode, appErr.GetMessage())

	// Convert ApplicationError to proto details using infrastructure utility
	protoDetails := ApplicationErrorToProtoDetails(appErr)

	stWithDetails, err := st.WithDetails(protoDetails)
	if err != nil {
		// Fallback if details can't be marshaled
		return st.Err()
	}

	return stWithDetails.Err()
}
