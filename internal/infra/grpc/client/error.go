package client

import (
	"fmt"

	infra_error "erp.localhost/internal/infra/error"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapGRPCError converts gRPC errors to domain errors
func mapGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return infra_error.NotFound(infra_error.NotFoundResource, st.Message(), nil)
	case codes.AlreadyExists:
		return infra_error.Conflict(infra_error.ConflictDuplicateResource).WithError(st.Err())
	case codes.InvalidArgument:
		return infra_error.Validation(infra_error.ValidationInvalidValue).WithError(st.Err())
	case codes.PermissionDenied:
		fallthrough
	case codes.Unauthenticated:
		return infra_error.Auth(infra_error.AuthPermissionDenied).WithError(st.Err())
	case codes.Internal:
		return infra_error.Internal(infra_error.InternalGRPCError, fmt.Errorf(st.Message()))
	case codes.Unavailable:
		return infra_error.Internal(infra_error.InternalGRPCError, fmt.Errorf("service unavailable: %s", st.Message()))
	default:
		return infra_error.Internal(infra_error.InternalGRPCError, fmt.Errorf("grpc error: %s", st.Message()))
	}
}
