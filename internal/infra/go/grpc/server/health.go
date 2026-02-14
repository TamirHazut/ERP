package server

import (
	"erp.localhost/infra/logging/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// RegisterHealthCheck registers the standard gRPC health check service on the provided server.
// The health check is automatically set to SERVING status for overall server health.
// This follows the standard gRPC health checking protocol: https://github.com/grpc/grpc/blob/master/doc/health-checking.md
func RegisterHealthCheck(srv *grpc.Server, log logger.Logger) {
	healthServer := health.NewServer()

	// Set overall health status to SERVING (empty string service name = overall health)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register the health server
	grpc_health_v1.RegisterHealthServer(srv, healthServer)

	log.Info("gRPC health check enabled")
}
