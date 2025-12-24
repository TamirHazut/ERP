#!/bin/bash
# Bash script for generating proto files on Linux/Mac

set -e

# Configuration
MODULE="erp.localhost"
PROTO_OUT="internal"
PROTO_COMMON="proto/common"
PROTO_AUTH="internal/auth/proto"
PROTO_CONFIG="internal/config/proto"
PROTO_CORE="internal/core/proto"

SERVICE="${1:-all}"

generate_common() {
    echo "Generating common proto files..."
    if [ -f "$PROTO_COMMON/common.proto" ]; then
        protoc --go_out=$PROTO_OUT \
            --go_opt=module=$MODULE \
            --go-grpc_out=$PROTO_OUT \
            --go-grpc_opt=module=$MODULE \
            -I=proto \
            "$PROTO_COMMON/common.proto"
        echo "✓ Common proto files generated"
    else
        echo "⚠ No common.proto file found, skipping..."
    fi
}

generate_auth() {
    echo "Generating auth service proto files..."
    if [ -f "$PROTO_AUTH/auth.proto" ]; then
        protoc --go_out=$PROTO_OUT \
            --go_opt=module=$MODULE \
            --go-grpc_out=$PROTO_OUT \
            --go-grpc_opt=module=$MODULE \
            -I=proto \
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
            -I=proto \
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
            -I=proto \
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
        generate_common
        generate_auth
        generate_config
        generate_core
        ;;
    common)
        generate_common
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
        echo "Usage: $0 [all|common|auth|config|core]"
        exit 1
        ;;
esac

echo ""
echo "✓ Proto generation complete!"

