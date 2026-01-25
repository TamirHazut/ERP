"""
User repository for Auth service functional tests.
Direct MongoDB operations for test data setup/teardown.
"""
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../infra/functional'))

from db.mongo_client import MongoDBClient
from typing import Optional, List, Dict, Any


class UserRepository:
    """Repository for User collection operations."""

    COLLECTION = "users"

    def __init__(self, mongo_client: MongoDBClient):
        self.mongo = mongo_client

    def create(self, user_data: Dict[str, Any]) -> str:
        """Create a new user."""
        return self.mongo.insert_one(self.COLLECTION, user_data)

    def find_by_email(self, tenant_id: str, email: str) -> Optional[Dict[str, Any]]:
        """Find user by email."""
        return self.mongo.find_one(self.COLLECTION, {
            "tenant_id": tenant_id,
            "email": email
        })

    def find_by_id(self, user_id: str) -> Optional[Dict[str, Any]]:
        """Find user by ID."""
        return self.mongo.find_one(self.COLLECTION, {"_id": user_id})

    def delete_by_email(self, tenant_id: str, email: str) -> bool:
        """Delete user by email."""
        return self.mongo.delete_one(self.COLLECTION, {
            "tenant_id": tenant_id,
            "email": email
        })

    def delete_all(self) -> int:
        """Delete all users."""
        return self.mongo.delete_many(self.COLLECTION)
