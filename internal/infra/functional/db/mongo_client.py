"""
MongoDB client wrapper for functional tests.
Provides CRUD operations for test data management.
"""
from pymongo import MongoClient
from pymongo.collection import Collection
from pymongo.database import Database
from typing import Optional, List, Dict, Any
from ..config import TestConfig


class MongoDBClient:
    """MongoDB client for functional test data management."""

    def __init__(self, database_name: str = None):
        self.database_name = database_name or TestConfig.MONGODB.database
        self.client: Optional[MongoClient] = None
        self.db: Optional[Database] = None

    def connect(self):
        """Establish MongoDB connection."""
        self.client = MongoClient(TestConfig.MONGODB.mongo_uri)
        self.db = self.client[self.database_name]
        # Verify connection
        self.client.admin.command('ping')

    def disconnect(self):
        """Close MongoDB connection."""
        if self.client:
            self.client.close()

    def get_collection(self, collection_name: str) -> Collection:
        """Get a MongoDB collection."""
        if not self.db:
            raise RuntimeError("Database not connected. Call connect() first.")
        return self.db[collection_name]

    def insert_one(self, collection_name: str, document: Dict[str, Any]) -> str:
        """Insert a single document."""
        collection = self.get_collection(collection_name)
        result = collection.insert_one(document)
        return str(result.inserted_id)

    def insert_many(self, collection_name: str, documents: List[Dict[str, Any]]) -> List[str]:
        """Insert multiple documents."""
        collection = self.get_collection(collection_name)
        result = collection.insert_many(documents)
        return [str(id) for id in result.inserted_ids]

    def find_one(self, collection_name: str, filter_dict: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Find a single document."""
        collection = self.get_collection(collection_name)
        return collection.find_one(filter_dict)

    def find_many(self, collection_name: str, filter_dict: Dict[str, Any] = None) -> List[Dict[str, Any]]:
        """Find multiple documents."""
        collection = self.get_collection(collection_name)
        return list(collection.find(filter_dict or {}))

    def update_one(self, collection_name: str, filter_dict: Dict[str, Any], update_dict: Dict[str, Any]) -> bool:
        """Update a single document."""
        collection = self.get_collection(collection_name)
        result = collection.update_one(filter_dict, {"$set": update_dict})
        return result.modified_count > 0

    def delete_one(self, collection_name: str, filter_dict: Dict[str, Any]) -> bool:
        """Delete a single document."""
        collection = self.get_collection(collection_name)
        result = collection.delete_one(filter_dict)
        return result.deleted_count > 0

    def delete_many(self, collection_name: str, filter_dict: Dict[str, Any] = None) -> int:
        """Delete multiple documents."""
        collection = self.get_collection(collection_name)
        result = collection.delete_many(filter_dict or {})
        return result.deleted_count

    def drop_database(self):
        """Drop the entire test database."""
        if self.client:
            self.client.drop_database(self.database_name)

    def __enter__(self):
        """Context manager entry."""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.disconnect()
