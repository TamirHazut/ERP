"""
Database utilities for functional tests.
Handles test database setup, teardown, and cleanup.
"""
from pymongo import MongoClient
import redis
import sys
import os

# Add infra functional path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from config import TestConfig


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
            
    # TODO: fix problematic login since there are multiple db's and only the test itself knows which db he needs
    def clean_test_data(self):
        """Clean all test data from databases."""
        if self.mongo_client:
            # Drop all test databases from central config
            for db_name in TestConfig.TEST_DATABASES.values():
                self.mongo_client.drop_database(db_name)

        if self.redis_client:
            # Flush Redis (careful - only use in test environment!)
            self.redis_client.flushdb()

    def seed_system_data(self):
        """Seed minimum required system data (tenant, admin user, etc.)."""
        # This would typically call the seeder or insert directly
        # Implementation depends on whether seeder supports test DB
        pass
