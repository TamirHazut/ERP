# Proto Definitions

This directory contains shared protobuf definitions used across multiple services.

## Directory Structure

```
proto/
├── common/              # Shared types (errors, base messages, common structures)
│   └── common.proto
└── README.md           # This file
```

## Service-Specific Proto Files

Each service has its own proto directory:
- `internal/auth/proto/` - Auth service definitions
- `internal/config/proto/` - Config service definitions
- `internal/core/proto/` - Core service definitions
- `internal/gateway/proto/` - Gateway service definitions (if needed)
- `internal/events/proto/` - Events service definitions (if needed)

## Why This Structure?

This structure supports:
1. **Service Ownership**: Each service owns its proto definitions
2. **Easy Service Splitting**: When services are split into separate repos, each service directory (including its proto files) can be moved independently
3. **Clear Boundaries**: Services only import what they need
4. **Shared Types**: Common types live in `proto/common/` and can become a shared module later

## Code Generation

### Using Makefile (Linux/Mac)

```bash
# Generate all proto files
make proto

# Generate specific service proto files
make proto-common
make proto-auth
make proto-config
make proto-core
```

### Using PowerShell Script (Windows)

```powershell
# Generate all proto files
.\scripts\generate-proto.ps1

# Generate specific service proto files
.\scripts\generate-proto.ps1 -Service common
.\scripts\generate-proto.ps1 -Service auth
.\scripts\generate-proto.ps1 -Service config
.\scripts\generate-proto.ps1 -Service core
```

### Using Bash Script (Linux/Mac)

```bash
# Generate all proto files
./scripts/generate-proto.sh

# Generate specific service proto files
./scripts/generate-proto.sh common
./scripts/generate-proto.sh auth
./scripts/generate-proto.sh config
./scripts/generate-proto.sh core
```

### Manual Generation

If you prefer to run protoc manually:

```bash
# Generate common proto
protoc --go_out=internal \
  --go_opt=module=erp.localhost \
  --go-grpc_out=internal \
  --go-grpc_opt=module=erp.localhost \
  -I=proto \
  proto/common/common.proto

# Generate service proto (example: auth)
protoc --go_out=internal \
  --go_opt=module=erp.localhost \
  --go-grpc_out=internal \
  --go-grpc_opt=module=erp.localhost \
  -I=proto \
  -I=internal/auth/proto \
  internal/auth/proto/*.proto
```

## Generated Code Location

Generated Go code will be placed in:
- Common: `internal/proto/common/`
- Auth: `internal/auth/proto/` (generated files)
- Config: `internal/config/proto/` (generated files)
- Core: `internal/core/proto/` (generated files)

## Prerequisites

1. **protoc** - Protocol Buffer compiler
   - Install from: https://grpc.io/docs/protoc-installation/
   - Or: `choco install protoc` (Windows)

2. **protoc-gen-go** - Go code generator
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   ```

3. **protoc-gen-go-grpc** - gRPC Go code generator
   ```bash
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

## Proto File Template

When creating a new service proto file, use this template:

```protobuf
syntax = "proto3";

package <service_name>;

option go_package = "erp.localhost/internal/<service>/proto";

import "common/common.proto";

// Your service definitions here
service <ServiceName> {
  rpc MethodName(Request) returns (Response);
}
```
