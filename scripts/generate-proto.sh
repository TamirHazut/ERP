#!/bin/bash
# Bash script for generating proto files on Linux/Mac

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Change to project root (parent of scripts directory)
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Configuration
MODULE="erp.localhost"
PROTO_OUT="."  # Output to project root (protoc will use go_package path relative to module)
PROTO_COMMON="internal/infra/proto"
PROTO_AUTH="internal/auth/proto"
PROTO_CONFIG="internal/config/proto"
PROTO_CORE="internal/core/proto"

SERVICE="${1:-all}"

generate_infra() {
    echo "Generating infra proto files..."
    if [ -f "$PROTO_COMMON/infra.proto" ]; then
        protoc --go_out=$PROTO_OUT \
            --go_opt=module=$MODULE \
            --go-grpc_out=$PROTO_OUT \
            --go-grpc_opt=module=$MODULE \
            -I=$PROTO_COMMON \
            "$PROTO_COMMON/infra.proto"
        echo "✓ Common proto files generated"
    else
        echo "⚠ No infra.proto file found, skipping..."
    fi
}

generate_auth() {
    echo "Generating auth service proto files..."
    if [ -f "$PROTO_AUTH/auth.proto" ]; then
        protoc --go_out=$PROTO_OUT \
            --go_opt=module=$MODULE \
            --go-grpc_out=$PROTO_OUT \
            --go-grpc_opt=module=$MODULE \
            -I=$PROTO_COMMON \
            -I=$PROTO_AUTH \
            "$PROTO_AUTH"/*.proto
        echo "✓ Auth proto files generated"
    else
        echo "⚠ No auth.proto file found, skipping..."
    fi
}

generate_config() {
    echo "Generating config service proto files..."
    if [ -f "$PROTO_CONFIG/config.proto" ]; then
        protoc --go_out=$PROTO_OUT \
            --go_opt=module=$MODULE \
            --go-grpc_out=$PROTO_OUT \
            --go-grpc_opt=module=$MODULE \
            -I=$PROTO_COMMON \
            -I=$PROTO_CONFIG \
            "$PROTO_CONFIG"/*.proto
        echo "✓ Config proto files generated"
    else
        echo "⚠ No config.proto file found, skipping..."
    fi
}

generate_core() {
    echo "Generating core service proto files..."
    if [ -f "$PROTO_CORE/core.proto" ]; then
        protoc --go_out=$PROTO_OUT \
            --go_opt=module=$MODULE \
            --go-grpc_out=$PROTO_OUT \
            --go-grpc_opt=module=$MODULE \
            -I=$PROTO_COMMON \
            -I=$PROTO_CORE \
            "$PROTO_CORE"/*.proto
        echo "✓ Core proto files generated"
    else
        echo "⚠ No core.proto file found, skipping..."
    fi
}

echo "=== Proto Code Generation ==="

case "$SERVICE" in
    all)
        generate_infra
        generate_auth
        generate_config
        generate_core
        ;;
    infra)
        generate_infra
        ;;
    auth)
        generate_auth
        ;;
    config)
        generate_config
        ;;
    core)
        generate_core
        ;;
    *)
        echo "Usage: $0 [all|infra|auth|config|core]"
        exit 1
        ;;
esac

echo ""
echo "✓ Proto generation complete!"

