"""
System data seeder for functional tests.
Seeds minimum required data (tenant, admin user, roles).
Permissions are code-defined in the registry — no DB documents needed.
"""
from datetime import datetime, UTC
from typing import Dict, Any
import bcrypt
import sys
import os

# Add infra functional path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from db.mongo_client import MongoDBClient
from db.redis_client import RedisClient
from config import TestConfig
from logger import get_logger

# Module logger
logger = get_logger("seeders.system")


class SystemSeeder:
    """Seeds system-level test data."""
    
    def __init__(self, mongo_client: MongoDBClient, redis_client: RedisClient = None):
        self.mongo = mongo_client
        self.redis = redis_client

    def seed_indexes(self):
        """Create MongoDB indexes for all collections."""
        logger.info("Creating indexes for system collections")

        # Tenants indexes
        self.mongo.create_index("tenants", [("name", 1)], unique=True, name="idx_name_unique")
        self.mongo.create_index("tenants", [("status", 1)], name="idx_status")
        self.mongo.create_index("tenants", [("domain", 1)], sparse=True, name="idx_domain")
        logger.debug("Created indexes for tenants collection")

        # Users indexes
        self.mongo.create_index("users", [("tenant_id", 1), ("email", 1)], unique=True, name="idx_tenant_email_unique")
        self.mongo.create_index("users", [("tenant_id", 1), ("username", 1)], unique=True, name="idx_tenant_username_unique")
        self.mongo.create_index("users", [("tenant_id", 1)], name="idx_tenant_id")
        self.mongo.create_index("users", [("tenant_id", 1), ("status", 1)], name="idx_tenant_status")
        self.mongo.create_index("users", [("tenant_id", 1), ("roles.role_id", 1)], name="idx_tenant_roles")
        logger.debug("Created indexes for users collection")

        # Roles indexes
        self.mongo.create_index("roles", [("tenant_id", 1), ("name", 1)], unique=True, name="idx_tenant_name_unique")
        self.mongo.create_index("roles", [("tenant_id", 1)], name="idx_tenant_id")
        self.mongo.create_index("roles", [("tenant_id", 1), ("permissions", 1)], name="idx_tenant_permissions")
        logger.debug("Created indexes for roles collection")

        logger.info("All indexes created successfully")

    def seed_all(self) -> Dict[str, str]:
        """
        Seed all system data.
        Returns: Dictionary with IDs of created entities.
        """
        logger.info("Starting system data seeding")

        # Create indexes first
        self.seed_indexes()

        # Create tenant
        tenant_id = self.seed_tenant()

        # Create system role with wildcard permission string
        role_id = self.seed_role(tenant_id, "*:*")

        # Create admin user
        user_id = self.seed_admin_user(tenant_id, role_id)

        logger.info(f"System seeding completed: tenant_id={tenant_id}, role_id={role_id}, user_id={user_id}")

        if self.redis:
            try:
                self.redis.set("system:tenant", tenant_id)
                self.redis.set("system:role", role_id)
                self.redis.set("system:user", user_id)
                logger.info(f"System IDs written to Redis")
            except Exception as e:
                logger.warning(f"Failed to write system IDs to Redis: {e}")

        return {
            "tenant_id": tenant_id,
            "role_id": role_id,
            "user_id": user_id,
        }

    def seed_tenant(self) -> str:
        """Seed test tenant."""
        tenant = {
            "name": TestConfig.DEFAULT_TENANT_NAME,
            "slug": "test-tenant",
            "status": 1,  # ACTIVE
            "created_at": datetime.now(),
            "created_by": "System",
            "protected": True
        }
        tenant_id = self.mongo.insert_one("tenants", tenant)
        logger.debug(f"Created tenant: id={tenant_id}, name={TestConfig.DEFAULT_TENANT_NAME}")
        return tenant_id

    def seed_role(self, tenant_id: str, permission_string: str) -> str:
        """Seed system admin role with the given permission string (e.g. '*:*')."""
        role = {
            "tenant_id": tenant_id,
            "name": "system_admin",
            "description": "System administrator role",
            "permissions": [permission_string],
            "status": 1,  # ACTIVE
            "type": 0,  # SYSTEM
            "created_at": datetime.now(),
            "created_by": "System",
            "protected": True
        }
        role_id = self.mongo.insert_one("roles", role)
        logger.debug(f"Created role: id={role_id}, name=system_admin, permissions=[{permission_string}]")
        return role_id

    def seed_admin_user(self, tenant_id: str, role_id: str) -> str:
        """Seed system admin user."""
        # Hash password
        password_hash = bcrypt.hashpw(
            TestConfig.DEFAULT_ADMIN_PASSWORD.encode('utf-8'),
            bcrypt.gensalt()
        ).decode('utf-8')

        user = {
            "tenant_id": tenant_id,
            "email": TestConfig.DEFAULT_ADMIN_EMAIL,
            "username": TestConfig.DEFAULT_ADMIN_USERNAME,
            "password": password_hash,
            "status": 1,  # ACTIVE
            "email_verified": True,
            "roles": [{
                "role_id": role_id,
                "tenant_id": tenant_id,
                "assigned_at": datetime.now(),
                "assigned_by": "System"
            }],
            "created_at": datetime.now(),
            "created_by": "System",
            "protected": True
        }
        user_id = self.mongo.insert_one("users", user)
        logger.debug(f"Created user: id={user_id}, email={TestConfig.DEFAULT_ADMIN_EMAIL}, roles=[{role_id}]")
        return user_id
