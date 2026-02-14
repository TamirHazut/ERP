# Functional Test Infrastructure

This directory contains the shared Python-based functional testing framework for the ERP system. The tests validate end-to-end service flows through black-box testing using gRPC.

## Overview

The functional test infrastructure is **completely independent** from Go production code and validates services through their public gRPC APIs. All test utilities, database clients, and gRPC clients are reimplemented in Python.

## Directory Structure

```
internal/infra/functional/
├── __init__.py
├── requirements.txt         # Python dependencies
├── pytest.ini              # Pytest configuration
├── config.py               # Test configuration (services, databases)
├── grpc_client.py          # Generic gRPC client wrapper
├── db/                     # Database clients
│   ├── mongo_client.py     # MongoDB operations
│   ├── redis_client.py     # Redis operations
│   └── manager.py          # Database lifecycle management
├── seeders/                # Test data seeders
│   └── system_seeder.py    # System-level data seeder
├── fixtures/               # Shared test fixtures
└── proto/                  # Generated Python proto files (auto-generated)
    ├── auth/v1/
    ├── config/v1/
    ├── core/v1/
    └── infra/v1/
```

## Quick Start

### 1. Generate Python Proto Files

```bash
make proto-python
```

This generates Python gRPC stubs from all proto definitions in `internal/infra/proto/`.

### 2. Install Dependencies

```bash
make test-functional-setup
```

This installs all required Python packages from `requirements.txt`.

### 3. Run Tests

```bash
# Run tests for a specific service
make test-functional-auth

# Run all functional tests
make test-functional-all

# Clean test artifacts
make test-functional-clean
```

## Configuration

Test configuration is centralized in `config.py`:

- **Service Endpoints**: Auth (port 5000), Config (port 5002), Core (port 5001)
- **MongoDB**: Test database `auth_db_test` (isolated from development)
- **Redis**: Same instance as development, flushed before/after tests
- **Test Credentials**: Default admin user credentials for authentication

## Database Isolation

Tests use a **separate test database** (`auth_db_test`) to avoid polluting development data. The database is automatically cleaned before and after test runs.

## Writing Tests

### Service-Specific Tests

Each service has its own `functional/` directory:

```
internal/<service>/functional/
├── conftest.py                 # Service-specific fixtures
├── clients/                    # gRPC client wrappers
├── repositories/               # Direct DB operations
├── seeders/                    # Service-specific seeders
└── test_*.py                   # Test files
```

### Example Test

```python
import pytest
import sys
import os

# Add infra functional path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../infra/functional'))

from grpc_client import GrpcClient
from config import TestConfig
from proto.auth.v1 import auth_pb2, auth_pb2_grpc

@pytest.mark.auth
class TestAuthenticationFlows:
    def test_login_success(self, clean_database):
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())
            request = auth_pb2.LoginRequest(
                tenant_id="test-tenant",
                account_id="user@example.com",
                password="password"
            )
            response = stub.Login(request)
            assert response.tokens.token != ""
```

## Components

### GrpcClient

Generic gRPC client wrapper with context manager support:

```python
from grpc_client import GrpcClient
from config import TestConfig

with GrpcClient(TestConfig.AUTH_SERVICE) as client:
    channel = client.get_channel()
    # Use channel for gRPC calls
```

### MongoDBClient

MongoDB operations for test data management:

```python
from db.mongo_client import MongoDBClient

with MongoDBClient() as mongo:
    # Insert test data
    user_id = mongo.insert_one("users", {"email": "test@example.com"})

    # Query data
    user = mongo.find_one("users", {"_id": user_id})

    # Cleanup
    mongo.delete_many("users")
```

### RedisClient

Redis operations for token/session testing:

```python
from db.redis_client import RedisClient

with RedisClient() as redis:
    redis.set("key", "value", ex=3600)
    value = redis.get("key")
    redis.delete("key")
```

### SystemSeeder

Seeds minimum required system data (tenant, admin user, roles, permissions):

```python
from db.mongo_client import MongoDBClient
from seeders.system_seeder import SystemSeeder

with MongoDBClient() as mongo:
    seeder = SystemSeeder(mongo)
    ids = seeder.seed_all()
    # Returns: {"tenant_id": "...", "user_id": "...", ...}
```

## Fixtures

### Global Fixtures (conftest.py)

- `setup_test_environment` (session) - Sets up databases once per session
- `clean_database` (function) - Cleans database before each test
- `test_config` (session) - Provides TestConfig instance

## Best Practices

1. **Test Isolation**: Each test should clean up its own data or use `clean_database` fixture
2. **Black-Box Testing**: Tests should only call gRPC APIs, not internal Go code
3. **Happy Path First**: Focus on successful flows initially
4. **Descriptive Names**: Use clear test names (e.g., `test_login_success`)
5. **Setup/Act/Assert**: Follow the standard test pattern

## Makefile Targets

```bash
# Proto generation
make proto-python              # Generate Python proto files
make proto-python-clean        # Clean generated proto files

# Test execution
make test-functional-setup     # Install Python dependencies
make test-functional-auth      # Run Auth service tests
make test-functional-config    # Run Config service tests
make test-functional-core      # Run Core service tests
make test-functional-all       # Run all service tests
make test-functional-clean     # Clean test artifacts
```

## Troubleshooting

### Import Errors

If you get import errors for proto files, ensure:
1. Proto files are generated: `make proto-python`
2. Python path is set correctly in your test files

### Database Connection Errors

Ensure MongoDB and Redis are running:
```bash
make docker-up
```

### Service Not Running

Tests require services to be running:
```bash
make run-auth
```

## Future Enhancements

- [ ] Service health checks before test execution
- [ ] Test data factories with Faker
- [ ] Negative test cases (error scenarios)
- [ ] Performance testing with timing assertions
- [ ] mTLS support when certificates are implemented
- [ ] CI/CD integration
