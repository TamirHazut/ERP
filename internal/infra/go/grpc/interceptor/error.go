package interceptor

import (
	"context"
	"fmt"

	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ServerErrorInterceptor creates a server-side error handling interceptor.
// It provides a safety net that:
// - Catches and recovers from panics
// - Converts AppErrors to proper gRPC status errors
// - Wraps plain errors as Internal errors
// - Passes through existing gRPC status errors unchanged
func ServerErrorInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered in gRPC handler",
					"method", info.FullMethod,
					"panic", r,
				)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		// Call the handler
		resp, err = handler(ctx, req)

		// No error, pass through
		if err == nil {
			return resp, nil
		}

		// Check if it's already a gRPC status error
		if st, ok := status.FromError(err); ok && st.Code() != codes.OK {
			// Already a proper gRPC status error, pass through unchanged
			return resp, err
		}

		// Convert to gRPC error via ToGRPCError
		// This handles both *AppError and plain errors
		grpcErr := infra_error.ToGRPCError(err)
		if grpcErr != nil {
			return resp, grpcErr
		}

		// Fallback (should never happen if ToGRPCError is working correctly)
		log.Error("unexpected nil from ToGRPCError",
			"method", info.FullMethod,
			"original_error", err,
		)
		return resp, status.Error(codes.Internal, fmt.Sprintf("error: %v", err))
	}
}
