"""
Redis client wrapper for functional tests.
Provides key-value operations for token and session management testing.
"""
import redis
from typing import Optional, Any
import sys
import os

# Add infra functional path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from config import TestConfig
from logger import get_logger

# Module logger
logger = get_logger("db.redis")


class RedisClient:
    """Redis client for functional test data management."""

    def __init__(self):
        self.client: Optional[redis.Redis] = None

    def connect(self):
        """Establish Redis connection."""
        try:
            self.client = redis.Redis(
                host=TestConfig.REDIS.host,
                port=TestConfig.REDIS.port,
                password=TestConfig.REDIS.password,
                decode_responses=True
            )
            # Verify connection
            self.client.ping()
            logger.info(f"Connected to Redis: {TestConfig.REDIS.host}:{TestConfig.REDIS.port}")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            raise

    def disconnect(self):
        """Close Redis connection."""
        if self.client:
            self.client.close()
            logger.info("Disconnected from Redis")

    def set(self, key: str, value: str, ex: Optional[int] = None) -> bool:
        """Set a key-value pair with optional expiration."""
        if not self.client:
            raise RuntimeError("Redis not connected. Call connect() first.")
        result = self.client.set(key, value, ex=ex)
        ttl_info = f", ttl={ex}" if ex else ""
        logger.debug(f"set: key={key}{ttl_info}")
        return result

    def get(self, key: str) -> Optional[str]:
        """Get a value by key."""
        if not self.client:
            raise RuntimeError("Redis not connected. Call connect() first.")
        result = self.client.get(key)
        logger.debug(f"get: key={key}, found={result is not None}")
        return result

    def delete(self, *keys: str) -> int:
        """Delete one or more keys."""
        if not self.client:
            raise RuntimeError("Redis not connected. Call connect() first.")
        result = self.client.delete(*keys)
        logger.debug(f"delete: keys={keys}, deleted={result}")
        return result

    def exists(self, key: str) -> bool:
        """Check if a key exists."""
        if not self.client:
            raise RuntimeError("Redis not connected. Call connect() first.")
        result = self.client.exists(key) > 0
        logger.debug(f"exists: key={key}, exists={result}")
        return result

    def flushdb(self):
        """Flush all keys from current database."""
        if self.client:
            self.client.flushdb()
            logger.info("Flushed all keys from Redis database")

    def __enter__(self):
        """Context manager entry."""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.disconnect()
