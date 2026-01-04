package error

import (
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// categoryToGRPCCode maps error categories to gRPC status codes
var categoryToGRPCCode = map[ErrorCategory]codes.Code{
	CategoryAuth:       codes.Unauthenticated,
	CategoryValidation: codes.InvalidArgument,
	CategoryNotFound:   codes.NotFound,
	CategoryConflict:   codes.AlreadyExists,
	CategoryBusiness:   codes.FailedPrecondition,
	CategoryInternal:   codes.Internal,
}

// Special cases where AUTH errors map to PermissionDenied
var permissionDeniedCodes = map[string]bool{
	"AUTH_PERMISSION_DENIED":    true,
	"AUTH_INSUFFICIENT_ROLE":    true,
	"AUTH_TENANT_ACCESS_DENIED": true,
}

// ToGRPCError converts an AppError to a gRPC status error
func ToGRPCError(err *AppError) error {
	if err == nil {
		return nil
	}

	// Determine the gRPC status code
	grpcCode := codes.Internal
	if code, ok := categoryToGRPCCode[err.Category]; ok {
		grpcCode = code
	}

	// Special handling for permission-related auth errors
	if err.Category == CategoryAuth && permissionDeniedCodes[err.Code] {
		grpcCode = codes.PermissionDenied
	}

	// Create gRPC status with error details
	st := status.New(grpcCode, err.Message)

	// Add error details as JSON in the status details
	details := &errorDetails{
		Code:     err.Code,
		Category: string(err.Category),
		Message:  err.Message,
		Details:  err.Details,
	}

	// Serialize details to JSON and add to status
	if detailsJSON, jsonErr := json.Marshal(details); jsonErr == nil {
		st, _ = st.WithDetails(&errorDetailsProto{
			Json: string(detailsJSON),
		})
	}

	return st.Err()
}

// FromGRPCError extracts an AppError from a gRPC error
func FromGRPCError(err error) *AppError {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error, wrap it as an internal error
		return &AppError{
			Code:     InternalUnexpectedError.Code,
			Message:  err.Error(),
			Category: CategoryInternal,
			Details:  make(map[string]any),
			Err:      err,
		}
	}

	// Try to extract error details from status
	for _, detail := range st.Details() {
		if ed, ok := detail.(*errorDetailsProto); ok {
			var details errorDetails
			if jsonErr := json.Unmarshal([]byte(ed.Json), &details); jsonErr == nil {
				return &AppError{
					Code:     details.Code,
					Message:  details.Message,
					Category: ErrorCategory(details.Category),
					Details:  details.Details,
				}
			}
		}
	}

	// Fallback: create AppError from gRPC status code and message
	category := grpcCodeToCategory(st.Code())
	return &AppError{
		Code:     grpcCodeToErrorCode(st.Code()),
		Message:  st.Message(),
		Category: category,
		Details:  make(map[string]any),
	}
}

// grpcCodeToCategory maps gRPC status codes to error categories
func grpcCodeToCategory(code codes.Code) ErrorCategory {
	switch code {
	case codes.Unauthenticated, codes.PermissionDenied:
		return CategoryAuth
	case codes.InvalidArgument:
		return CategoryValidation
	case codes.NotFound:
		return CategoryNotFound
	case codes.AlreadyExists:
		return CategoryConflict
	case codes.FailedPrecondition:
		return CategoryBusiness
	default:
		return CategoryInternal
	}
}

// grpcCodeToErrorCode generates a generic error code from gRPC status code
func grpcCodeToErrorCode(code codes.Code) string {
	switch code {
	case codes.Unauthenticated:
		return "AUTH_UNAUTHENTICATED"
	case codes.PermissionDenied:
		return "AUTH_PERMISSION_DENIED"
	case codes.InvalidArgument:
		return "VALIDATION_INVALID_ARGUMENT"
	case codes.NotFound:
		return "NOT_FOUND_RESOURCE"
	case codes.AlreadyExists:
		return "CONFLICT_RESOURCE_EXISTS"
	case codes.FailedPrecondition:
		return "BUSINESS_PRECONDITION_FAILED"
	case codes.Unavailable:
		return "INTERNAL_SERVICE_UNAVAILABLE"
	case codes.DeadlineExceeded:
		return "INTERNAL_TIMEOUT"
	default:
		return "INTERNAL_UNEXPECTED_ERROR"
	}
}

// errorDetails is used for JSON serialization of error details
type errorDetails struct {
	Code     string         `json:"code"`
	Category string         `json:"category"`
	Message  string         `json:"message"`
	Details  map[string]any `json:"details,omitempty"`
}

// errorDetailsProto is a simple proto-like struct for gRPC details
// This implements the proto.Message interface minimally for gRPC status details
type errorDetailsProto struct {
	Json string
}

func (e *errorDetailsProto) Reset()         { *e = errorDetailsProto{} }
func (e *errorDetailsProto) String() string { return e.Json }
func (e *errorDetailsProto) ProtoMessage()  {}

// GetGRPCCode returns the gRPC status code for an AppError
func GetGRPCCode(err *AppError) codes.Code {
	if err == nil {
		return codes.OK
	}

	grpcCode := codes.Internal
	if code, ok := categoryToGRPCCode[err.Category]; ok {
		grpcCode = code
	}

	if err.Category == CategoryAuth && permissionDeniedCodes[err.Code] {
		grpcCode = codes.PermissionDenied
	}

	return grpcCode
}

// IsGRPCError checks if an error is a gRPC status error
func IsGRPCError(err error) bool {
	_, ok := status.FromError(err)
	return ok
}
