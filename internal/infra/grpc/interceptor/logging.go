package interceptor

import (
	"context"
	"time"

	"erp.localhost/internal/infra/logging/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Shared logging helper
func logGRPCCall(log logger.Logger, method string, duration time.Duration, err error, isClient bool) {
	side := "server"
	if isClient {
		side = "client"
	}

	fields := []interface{}{
		"side", side,
		"method", method,
		"duration", duration,
	}

	if err != nil {
		st, _ := status.FromError(err)
		fields = append(fields, "error", err, "code", st.Code())
		log.Error("gRPC call failed", fields...)
	} else {
		log.Debug("gRPC call completed", fields...)
	}
}

// ClientLoggingInterceptor creates a client-side logging interceptor
func ClientLoggingInterceptor(log logger.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		log.Debug("gRPC client request started", "method", method)

		err := invoker(ctx, method, req, reply, cc, opts...)

		logGRPCCall(log, method, time.Since(start), err, true)
		return err
	}
}

// ServerLoggingInterceptor creates a server-side logging interceptor
func ServerLoggingInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		log.Debug("gRPC server request started", "method", info.FullMethod)

		resp, err := handler(ctx, req)

		logGRPCCall(log, info.FullMethod, time.Since(start), err, false)
		return resp, err
	}
}
