"""
MongoDB client wrapper for functional tests.
Provides CRUD operations for test data management.
"""
from pymongo import MongoClient
from pymongo.collection import Collection
from pymongo.database import Database
from typing import Optional, List, Dict, Any
import sys
import os

# Add infra functional path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from config import TestConfig
from logger import get_logger

# Module logger
logger = get_logger("db.mongo")


class MongoDBClient:
    """MongoDB client for functional test data management."""

    def __init__(self, database_name):
        self.database_name = database_name
        self.client: Optional[MongoClient] = None
        self.db: Optional[Database] = None

    def connect(self):
        """Establish MongoDB connection."""
        try:
            self.client = MongoClient(TestConfig.MONGODB.mongo_uri)
            self.db = self.client[self.database_name]
            # Verify connection
            self.client.admin.command('ping')
            logger.info(f"Connected to MongoDB: {TestConfig.MONGODB.mongo_uri}, database: {self.database_name}")
        except Exception as e:
            logger.error(f"Failed to connect to MongoDB: {e}")
            raise

    def disconnect(self):
        """Close MongoDB connection."""
        if self.client is not None:
            self.client.close()
            logger.info(f"Disconnected from MongoDB: database: {self.database_name}")

    def get_collection(self, collection_name: str) -> Collection:
        """Get a MongoDB collection."""
        if self.db is None:
            raise RuntimeError("Database not connected. Call connect() first.")
        return self.db[collection_name]

    def insert_one(self, collection_name: str, document: Dict[str, Any]) -> str:
        """Insert a single document."""
        collection = self.get_collection(collection_name)
        result = collection.insert_one(document)
        logger.debug(f"insert_one: collection={collection_name}, database={self.database_name}, id={result.inserted_id}")
        return str(result.inserted_id)

    def insert_many(self, collection_name: str, documents: List[Dict[str, Any]]) -> List[str]:
        """Insert multiple documents."""
        collection = self.get_collection(collection_name)
        result = collection.insert_many(documents)
        logger.debug(f"insert_many: collection={collection_name}, database={self.database_name}, count={len(result.inserted_ids)}")
        return [str(id) for id in result.inserted_ids]

    def find_one(self, collection_name: str, filter_dict: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Find a single document."""
        collection = self.get_collection(collection_name)
        result = collection.find_one(filter_dict)
        logger.debug(f"find_one: collection={collection_name}, database={self.database_name}, found={result is not None}")
        return result

    def find_many(self, collection_name: str, filter_dict: Dict[str, Any] = None) -> List[Dict[str, Any]]:
        """Find multiple documents."""
        collection = self.get_collection(collection_name)
        results = list(collection.find(filter_dict or {}))
        logger.debug(f"find_many: collection={collection_name}, database={self.database_name}, count={len(results)}")
        return results

    def update_one(self, collection_name: str, filter_dict: Dict[str, Any], update_dict: Dict[str, Any]) -> bool:
        """Update a single document."""
        collection = self.get_collection(collection_name)
        result = collection.update_one(filter_dict, {"$set": update_dict})
        logger.debug(f"update_one: collection={collection_name}, database={self.database_name}, modified={result.modified_count}")
        return result.modified_count > 0

    def delete_one(self, collection_name: str, filter_dict: Dict[str, Any]) -> bool:
        """Delete a single document."""
        collection = self.get_collection(collection_name)
        result = collection.delete_one(filter_dict)
        logger.debug(f"delete_one: collection={collection_name}, database={self.database_name}, deleted={result.deleted_count}")
        return result.deleted_count > 0

    def delete_many(self, collection_name: str, filter_dict: Dict[str, Any] = None) -> int:
        """Delete multiple documents."""
        collection = self.get_collection(collection_name)
        result = collection.delete_many(filter_dict or {})
        logger.debug(f"delete_many: collection={collection_name}, database={self.database_name}, deleted={result.deleted_count}")
        return result.deleted_count

    def create_index(self, collection_name: str, keys: List[tuple], unique: bool = False, sparse: bool = False, name: str = None):
        """Create a single index on a collection.

        Args:
            collection_name: Name of the collection
            keys: List of (field, direction) tuples, e.g., [("name", 1)] or [("tenant_id", 1), ("email", 1)]
            unique: Whether the index should enforce uniqueness
            sparse: Whether the index should be sparse
            name: Optional custom name for the index
        """
        collection = self.get_collection(collection_name)
        index_options = {"unique": unique, "sparse": sparse}
        if name:
            index_options["name"] = name

        collection.create_index(keys, **index_options)
        logger.debug(f"create_index: collection={collection_name}, keys={keys}, unique={unique}, sparse={sparse}, name={name}")

    def drop_database(self):
        """Drop the entire test database."""
        if self.client is not None:
            self.client.drop_database(self.database_name)
            logger.info(f"Dropped database: {self.database_name}")

    def __enter__(self):
        """Context manager entry."""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.disconnect()
