"""
Redis client wrapper for functional tests.
Provides key-value operations for token and session management testing.
"""
import redis
from typing import Optional, Any
from ..config import TestConfig


class RedisClient:
    """Redis client for functional test data management."""

    def __init__(self):
        self.client: Optional[redis.Redis] = None

    def connect(self):
        """Establish Redis connection."""
        self.client = redis.Redis(
            host=TestConfig.REDIS.host,
            port=TestConfig.REDIS.port,
            password=TestConfig.REDIS.password,
            decode_responses=True
        )
        # Verify connection
        self.client.ping()

    def disconnect(self):
        """Close Redis connection."""
        if self.client:
            self.client.close()

    def set(self, key: str, value: str, ex: Optional[int] = None) -> bool:
        """Set a key-value pair with optional expiration."""
        if not self.client:
            raise RuntimeError("Redis not connected. Call connect() first.")
        return self.client.set(key, value, ex=ex)

    def get(self, key: str) -> Optional[str]:
        """Get a value by key."""
        if not self.client:
            raise RuntimeError("Redis not connected. Call connect() first.")
        return self.client.get(key)

    def delete(self, *keys: str) -> int:
        """Delete one or more keys."""
        if not self.client:
            raise RuntimeError("Redis not connected. Call connect() first.")
        return self.client.delete(*keys)

    def exists(self, key: str) -> bool:
        """Check if a key exists."""
        if not self.client:
            raise RuntimeError("Redis not connected. Call connect() first.")
        return self.client.exists(key) > 0

    def flushdb(self):
        """Flush all keys from current database."""
        if self.client:
            self.client.flushdb()

    def __enter__(self):
        """Context manager entry."""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.disconnect()
