"""
Test configuration for functional tests.
Centralized configuration for service endpoints, database connections, and test data.
"""
import os
from dataclasses import dataclass
from typing import Optional


@dataclass
class ServiceConfig:
    """Configuration for a gRPC service."""
    name: str
    host: str
    port: int
    use_tls: bool = False

    @property
    def endpoint(self) -> str:
        return f"{self.host}:{self.port}"


@dataclass
class DatabaseConfig:
    """Database connection configuration."""
    host: str
    port: int
    username: str
    password: str
    database: str

    @property
    def mongo_uri(self) -> str:
        return f"mongodb://{self.username}:{self.password}@{self.host}:{self.port}"

    @property
    def redis_uri(self) -> str:
        return f"redis://:{self.password}@{self.host}:{self.port}"


class TestConfig:
    """Global test configuration."""

    # Service endpoints
    AUTH_SERVICE = ServiceConfig(
        name="auth",
        host=os.getenv("AUTH_SERVICE_HOST", "localhost"),
        port=int(os.getenv("AUTH_SERVICE_PORT", "5000")),
        use_tls=False  # Change to True when mTLS is implemented
    )

    CONFIG_SERVICE = ServiceConfig(
        name="config",
        host=os.getenv("CONFIG_SERVICE_HOST", "localhost"),
        port=int(os.getenv("CONFIG_SERVICE_PORT", "5002")),
        use_tls=False
    )

    CORE_SERVICE = ServiceConfig(
        name="core",
        host=os.getenv("CORE_SERVICE_HOST", "localhost"),
        port=int(os.getenv("CORE_SERVICE_PORT", "5001")),
        use_tls=False
    )

    # MongoDB configuration (TEST database)
    MONGODB = DatabaseConfig(
        host=os.getenv("MONGO_HOST", "localhost"),
        port=int(os.getenv("MONGO_PORT", "27017")),
        username=os.getenv("MONGO_USER", "root"),
        password=os.getenv("MONGO_PASSWORD", "secret"),
        database="auth_db_test"  # Separate test database
    )

    # Redis configuration
    REDIS = DatabaseConfig(
        host=os.getenv("REDIS_HOST", "localhost"),
        port=int(os.getenv("REDIS_PORT", "6379")),
        username="",
        password=os.getenv("REDIS_PASSWORD", "supersecretredis"),
        database="0"
    )

    # Test data defaults
    DEFAULT_TENANT_NAME = "test_tenant"
    DEFAULT_ADMIN_EMAIL = "admin@test.com"
    DEFAULT_ADMIN_PASSWORD = "TestPassword123!"
    DEFAULT_ADMIN_USERNAME = "test_admin"

    # Test timeouts
    GRPC_TIMEOUT = 10  # seconds
    DB_TIMEOUT = 5  # seconds
