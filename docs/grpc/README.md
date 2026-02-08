# gRPC Infrastructure

This document describes the generic gRPC infrastructure layer in `internal/infra/grpc/`. This layer provides reusable server and client components, interceptors for common concerns, and utilities for service-to-service communication.

## Overview

The gRPC infrastructure provides:

- **Server Bootstrap** (`internal/infra/grpc/server/`) - Generic gRPC server with TLS/mTLS, graceful shutdown, keepalive configuration
- **Client Bootstrap** (`internal/infra/grpc/client/`) - Generic gRPC client with TLS/mTLS, connection management
- **Interceptors** (`internal/infra/grpc/interceptor/`) - Middleware for logging, error handling, and metrics
- **Health Checks** - Standard gRPC health check protocol (automatic)
- **Error Handling** - Standardized error conversion between AppError and gRPC status codes

## Server Setup

### Basic Server Creation

```go
import (
    "erp.localhost/internal/infra/grpc/server"
    "erp.localhost/internal/infra/grpc/interceptor"
    "erp.localhost/internal/infra/logging/logger"
    "erp.localhost/internal/infra/model/shared"
)

func main() {
    log := logger.NewBaseLogger(shared.ModuleAuth)

    // Optional: create metrics collector
    metrics := interceptor.NewMetricsCollector()

    // Detect certificates (can be nil for insecure mode)
    certs := shared.NewCerts()

    srv, err := server.NewGRPCServer(&server.Config{
        Port:             5000,
        Module:           shared.ModuleAuth,
        Insecure:         certs == nil,
        Certs:            certs,
        EnableReflection: true,
        KeepAliveTime:    30 * time.Second,
        KeepAliveTimeout: 10 * time.Second,
        Metrics:          metrics, // Optional - pass nil to disable metrics
    }, log)

    if err != nil {
        log.Fatal("failed to create server", "error", err)
    }

    // Register your services
    srv.RegisterService(&authv1.AuthService_ServiceDesc, authServiceImpl)

    // Start server
    quit := make(chan struct{})
    go func() {
        if err := srv.ListenAndServe(quit); err != nil {
            log.Error("server error", "error", err)
        }
    }()

    // Wait for signal, then graceful shutdown
    // ... signal handling code ...
    close(quit)
}
```

### Server Configuration Fields

| Field | Type | Description |
|---|---|---|
| `Port` | `int` | Port number to listen on |
| `Module` | `shared.Module` | Service module identifier (Auth, Core, Config, etc.) |
| `Insecure` | `bool` | If true, runs without TLS (development only) |
| `Certs` | `*shared.Certs` | Certificates for mTLS (nil if Insecure=true) |
| `EnableReflection` | `bool` | Enable gRPC reflection for tools like grpcurl |
| `KeepAliveTime` | `time.Duration` | Keepalive ping interval (0 = disabled) |
| `KeepAliveTimeout` | `time.Duration` | Keepalive ping timeout |
| `MaxConnectionIdle` | `time.Duration` | Max idle time before connection close |
| `MaxConnectionAge` | `time.Duration` | Max connection lifetime |
| `Metrics` | `*interceptor.MetricsCollector` | Optional metrics collector (nil = no metrics) |

## Client Setup

### Basic Client Creation

```go
import (
    "context"
    "erp.localhost/internal/infra/grpc/client"
    "erp.localhost/internal/infra/grpc/interceptor"
    "erp.localhost/internal/infra/logging/logger"
    "erp.localhost/internal/infra/model/shared"
    authv1 "erp.localhost/internal/infra/model/auth/v1"
)

func main() {
    log := logger.NewBaseLogger(shared.ModuleCore)
    ctx := context.Background()

    // Optional: create metrics collector
    metrics := interceptor.NewMetricsCollector()

    // Detect certificates
    certs := shared.NewCerts()

    grpcClient, err := client.NewGRPCClient(ctx, &client.Config{
        Address:        "localhost:5000",
        Module:         shared.ModuleCore,
        Insecure:       certs == nil,
        Certs:          certs,
        ConnectTimeout: 5 * time.Second,
        RequestTimeout: 10 * time.Second,
        Metrics:        metrics, // Optional - pass nil to disable metrics
    }, log)

    if err != nil {
        log.Fatal("failed to create client", "error", err)
    }
    defer grpcClient.Close()

    // Create service client stub
    authClient := authv1.NewAuthServiceClient(grpcClient.Conn())

    // Make RPC calls
    resp, err := authClient.Login(ctx, &authv1.LoginRequest{
        TenantId: "tenant-123",
        Email:    "user@example.com",
        Password: "password",
    })
}
```

### Client Configuration Fields

| Field | Type | Description |
|---|---|---|
| `Address` | `string` | Server address (e.g., "localhost:5000") |
| `Module` | `shared.Module` | Client module identifier |
| `Insecure` | `bool` | If true, connects without TLS (development only) |
| `Certs` | `*shared.Certs` | Client certificates for mTLS (nil if Insecure=true) |
| `ConnectTimeout` | `time.Duration` | Connection timeout |
| `RequestTimeout` | `time.Duration` | Per-request timeout |
| `Metrics` | `*interceptor.MetricsCollector` | Optional metrics collector (nil = no metrics) |

## Interceptor Chain

Interceptors run in a specific order to ensure correct behavior.

### Server Interceptor Order

Interceptors are configured as a chain. The **first** interceptor in the chain is the **outermost** (runs first on request, last on response):

```
Request → Error → Metrics → Logging → Handler
Response ← Error ← Metrics ← Logging ← Handler
```

1. **Error Interceptor** (outermost)
   - Catches panics and recovers gracefully
   - Converts `*AppError` to gRPC status codes
   - Wraps plain errors as Internal errors
   - Passes through existing gRPC status errors unchanged
   - **Always active** (no configuration needed)

2. **Metrics Interceptor**
   - Records request count, failure count, latency buckets
   - **Optional** - only active if `Config.Metrics` is non-nil
   - No-op if `Metrics` field is nil

3. **Logging Interceptor** (innermost)
   - Logs request start and completion with duration
   - Logs errors with gRPC status code
   - **Always active** (no configuration needed)

### Client Interceptor Order

```
Request → Metrics → Logging → RPC Call
Response ← Metrics ← Logging ← RPC Call
```

1. **Metrics Interceptor** - Records client-side metrics (optional)
2. **Logging Interceptor** - Logs outgoing requests

**Note:** No error interceptor on client side. Errors from the server arrive as gRPC status codes and are mapped to `*AppError` by the `mapGRPCError()` function in per-method handler code (see `client/auth.go` for examples).

## Health Check

The gRPC infrastructure automatically registers a standard health check service on all servers. No configuration needed.

### Using Health Check

```bash
# Using grpcurl (requires reflection enabled)
grpcurl -plaintext localhost:5000 grpc.health.v1.Health/Check

# Response
{
  "status": "SERVING"
}
```

### Health Check in Code

```go
import "google.golang.org/grpc/health/grpc_health_v1"

healthClient := grpc_health_v1.NewHealthClient(conn)
resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
    Service: "", // empty string = overall health
})
```

## Metrics

The metrics interceptor collects basic in-memory metrics:

- **Total Requests** - Count of all requests
- **Failed Requests** - Count of requests that returned errors
- **Latency Buckets** - Cumulative latency distribution in 8 buckets (1, 5, 10, 25, 50, 100, 250, 500 ms)

### Using Metrics

```go
// Create collector
metrics := interceptor.NewMetricsCollector()

// Pass to server/client config
srv, _ := server.NewGRPCServer(&server.Config{
    Metrics: metrics,
    // ... other fields
}, log)

// Read snapshot periodically
snapshot := metrics.Snapshot()
log.Info("metrics",
    "total_requests", snapshot.TotalRequests,
    "failed_requests", snapshot.FailedRequests,
)

// Latency buckets (cumulative)
for i, count := range snapshot.LatencyBuckets {
    threshold := []float64{1, 5, 10, 25, 50, 100, 250, 500}[i]
    log.Info("latency bucket",
        "threshold_ms", threshold,
        "count", count, // requests with latency <= threshold
    )
}
```

The metrics are **in-memory only** and reset on service restart. For production observability, consider exporting to Prometheus, StatsD, or similar.

## Service-to-Service Authentication & Authorization

The ERP system uses a layered security model:

### Transport Security: mTLS

Service-to-service communication is secured via **mutual TLS (mTLS)**:

- Each service has a certificate identifying its module (auth, core, config, etc.)
- Server verifies client certificate at TLS handshake
- Client verifies server certificate
- This proves **service identity** at the transport layer

**No JWT bearer tokens in gRPC metadata.** The auth channel is the TLS certificate, not application-level tokens.

### User Context: Proto Fields

User identity (who is the end-user making this request?) flows via **proto message fields**:

```protobuf
message UserIdentifier {
  string tenant_id = 1;
  string user_id = 2;
}

message SomeRequest {
  UserIdentifier identifier = 1;
  // ... other fields
}
```

- **Gateway** validates the JWT token (exp check, signature, etc.)
- Gateway extracts `tenant_id` and `user_id` from the token
- Gateway passes these in the proto request to backend services
- If gateway needs to verify the token is not revoked, it calls `Auth.VerifyToken` gRPC method

### Authorization: CheckPermissions

Services call the Auth service's `CheckPermissions` gRPC method to verify user permissions:

```go
// In Core service, before updating a user
resp, err := authClient.CheckPermissions(ctx, &authv1.CheckPermissionsRequest{
    Identifier: &infrav1.UserIdentifier{
        TenantId: req.Identifier.TenantId,
        UserId:   req.Identifier.UserId,
    },
    Permissions: []string{"user:write"},
})

if !resp.GetHasPermissions() {
    return nil, infra_error.Auth(infra_error.AuthPermissionDenied)
}
```

### Summary

| Layer | Mechanism | Purpose |
|---|---|---|
| **Transport Identity** | mTLS certificates | Proves which service is calling (auth, core, config, etc.) |
| **User Context** | Proto `UserIdentifier` fields | Identifies the end-user making the request |
| **Authorization** | `Auth.CheckPermissions` RPC | Verifies user has permission to perform action |

## Error Handling

All errors flow through the standardized AppError system:

### Server-Side Error Flow

```go
// In your service implementation
func (s *MyService) MyMethod(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    // Business logic
    user, err := s.userRepo.GetUser(req.UserId)
    if err != nil {
        // Return AppError - error interceptor converts to gRPC status
        return nil, infra_error.NotFound(infra_error.NotFoundResource, "user", req.UserId)
    }

    return &pb.Response{...}, nil
}
```

The **error interceptor** automatically:
- Converts `*AppError` to gRPC status via `ToGRPCError()`
- Embeds error details as JSON in status
- Maps error categories to gRPC codes (AUTH → Unauthenticated, NOT_FOUND → NotFound, etc.)

### Client-Side Error Flow

```go
// In client code
resp, err := client.MyMethod(ctx, req)
if err != nil {
    // Use mapGRPCError to convert back to AppError
    appErr := mapGRPCError(err)

    // Or check gRPC code directly
    if st, ok := status.FromError(err); ok {
        if st.Code() == codes.NotFound {
            // handle not found
        }
    }
}
```

See `internal/infra/error/grpc.go` for full error mapping details.

## Code Examples

### Complete Server Example

```go
package main

import (
    "os"
    "os/signal"
    "syscall"

    "erp.localhost/internal/infra/grpc/server"
    "erp.localhost/internal/infra/grpc/interceptor"
    "erp.localhost/internal/infra/logging/logger"
    "erp.localhost/internal/infra/model/shared"
    authv1 "erp.localhost/internal/infra/model/auth/v1"
)

func main() {
    log := logger.NewBaseLogger(shared.ModuleAuth)
    metrics := interceptor.NewMetricsCollector()

    srv, err := server.NewGRPCServer(&server.Config{
        Port:             5000,
        Module:           shared.ModuleAuth,
        Insecure:         false,
        Certs:            shared.NewCerts(),
        EnableReflection: true,
        Metrics:          metrics,
    }, log)
    if err != nil {
        log.Fatal("failed to create server", "error", err)
    }

    // Register services
    authService := NewAuthService(/* deps */, log)
    srv.RegisterService(&authv1.AuthService_ServiceDesc, authService)

    // Graceful shutdown
    quit := make(chan struct{})
    stopChan := make(chan os.Signal, 1)
    signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        if err := srv.ListenAndServe(quit); err != nil {
            log.Error("server stopped", "error", err)
        }
    }()

    <-stopChan
    log.Info("shutting down...")
    close(quit)
}
```

### Complete Client Example

```go
package main

import (
    "context"

    "erp.localhost/internal/infra/grpc/client"
    "erp.localhost/internal/infra/grpc/interceptor"
    "erp.localhost/internal/infra/logging/logger"
    "erp.localhost/internal/infra/model/shared"
    authv1 "erp.localhost/internal/infra/model/auth/v1"
    infrav1 "erp.localhost/internal/infra/model/infra/v1"
)

func main() {
    log := logger.NewBaseLogger(shared.ModuleCore)
    ctx := context.Background()

    grpcClient, err := client.NewGRPCClient(ctx, &client.Config{
        Address:  "localhost:5000",
        Module:   shared.ModuleCore,
        Insecure: false,
        Certs:    shared.NewCerts(),
        Metrics:  interceptor.NewMetricsCollector(),
    }, log)
    if err != nil {
        log.Fatal("failed to create client", "error", err)
    }
    defer grpcClient.Close()

    authClient := authv1.NewAuthServiceClient(grpcClient.Conn())

    // Make RPC call
    resp, err := authClient.CheckPermissions(ctx, &authv1.CheckPermissionsRequest{
        Identifier: &infrav1.UserIdentifier{
            TenantId: "tenant-123",
            UserId:   "user-456",
        },
        Permissions: []string{"user:read"},
    })

    if err != nil {
        log.Error("RPC failed", "error", err)
        return
    }

    log.Info("permission check", "has_permissions", resp.GetHasPermissions())
}
```

## Testing

All interceptors have comprehensive unit tests. Run them with:

```bash
# Test all gRPC infrastructure
go test ./internal/infra/grpc/...

# Test specific component
go test ./internal/infra/grpc/interceptor
go test ./internal/infra/grpc/server
go test ./internal/infra/grpc/client
```

## Future Enhancements

The following features are deferred for future implementation:

- Advanced connection pooling strategies
- Retry logic with exponential backoff
- Circuit breaker pattern
- Load balancing
- Service discovery integration
- Distributed tracing (OpenTelemetry)
- Prometheus metrics export
- Stream interceptors (currently only unary)
