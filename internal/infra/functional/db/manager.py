"""
Database utilities for functional tests.
Handles test database setup, teardown, and cleanup.
"""
from pymongo import MongoClient
import redis
from ..config import TestConfig


class DatabaseManager:
    """Manages test database lifecycle."""

    def __init__(self):
        self.mongo_client = None
        self.redis_client = None

    def setup(self):
        """Setup test databases."""
        # MongoDB connection
        self.mongo_client = MongoClient(TestConfig.MONGODB.mongo_uri)

        # Redis connection
        self.redis_client = redis.Redis(
            host=TestConfig.REDIS.host,
            port=TestConfig.REDIS.port,
            password=TestConfig.REDIS.password,
            decode_responses=True
        )

        # Verify connections
        self.mongo_client.admin.command('ping')
        self.redis_client.ping()

    def teardown(self):
        """Teardown database connections."""
        if self.mongo_client:
            self.mongo_client.close()
        if self.redis_client:
            self.redis_client.close()

    def clean_test_data(self):
        """Clean all test data from databases."""
        if self.mongo_client:
            # Drop test database
            db = self.mongo_client[TestConfig.MONGODB.database]
            db.client.drop_database(TestConfig.MONGODB.database)

        if self.redis_client:
            # Flush Redis (careful - only use in test environment!)
            self.redis_client.flushdb()

    def seed_system_data(self):
        """Seed minimum required system data (tenant, admin user, etc.)."""
        # This would typically call the seeder or insert directly
        # Implementation depends on whether seeder supports test DB
        pass
