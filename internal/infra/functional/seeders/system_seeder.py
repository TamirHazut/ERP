"""
System data seeder for functional tests.
Seeds minimum required data (tenant, admin user, roles, permissions).
"""
from datetime import datetime, UTC
from typing import Dict, Any
import bcrypt
import sys
import os

# Add infra functional path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from db.mongo_client import MongoDBClient
from config import TestConfig
from logger import get_logger

# Module logger
logger = get_logger("seeders.system")


class SystemSeeder:
    """Seeds system-level test data."""
    
    def __init__(self, mongo_client: MongoDBClient):
        self.mongo = mongo_client

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

        # Permissions indexes
        self.mongo.create_index("permissions", [("tenant_id", 1), ("permission_string", 1)], unique=True, name="idx_tenant_permission_unique")
        self.mongo.create_index("permissions", [("tenant_id", 1), ("resource", 1)], name="idx_tenant_resource")
        self.mongo.create_index("permissions", [("tenant_id", 1), ("resource", 1), ("action", 1)], name="idx_tenant_resource_action")
        logger.debug("Created indexes for permissions collection")

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

        # Create system permission
        permission_id = self.seed_permission(tenant_id)

        # Create system role
        role_id = self.seed_role(tenant_id, permission_id)

        # Create admin user
        user_id = self.seed_admin_user(tenant_id, role_id)

        logger.info(f"System seeding completed: tenant_id={tenant_id}, role_id={role_id}, permission_id={permission_id}, user_id={user_id}")

        return {
            "tenant_id": tenant_id,
            "permission_id": permission_id,
            "role_id": role_id,
            "user_id": user_id
        }

    def seed_tenant(self) -> str:
        """Seed test tenant."""
        tenant = {
            "name": TestConfig.DEFAULT_TENANT_NAME,
            "slug": "test-tenant",
            "status": 1,  # ACTIVE
            "created_at": datetime.now(),
            "created_by": "System"
        }
        tenant_id = self.mongo.insert_one("tenants", tenant)
        logger.debug(f"Created tenant: id={tenant_id}, name={TestConfig.DEFAULT_TENANT_NAME}")
        return tenant_id

    def seed_permission(self, tenant_id: str) -> str:
        """Seed system admin permission."""
        permission = {
            "tenant_id": tenant_id,
            "resource": "*",
            "action": "*",
            "permission_string": "*:*",
            "display_name": "System Controller",
            "description": "Full system access",
            "status": 1,  # ACTIVE
            "is_dangerous": True,
            "created_at": datetime.now(),
            "created_by": "System"
        }
        permission_id = self.mongo.insert_one("permissions", permission)
        logger.debug(f"Created permission: id={permission_id}, permission_string=*:*")
        return permission_id

    def seed_role(self, tenant_id: str, permission_id: str) -> str:
        """Seed system admin role."""
        role = {
            "tenant_id": tenant_id,
            "name": "system_admin",
            "description": "System administrator role",
            "permissions": [permission_id],
            "status": 1,  # ACTIVE
            "type": 0,  # SYSTEM
            "created_at": datetime.now(),
            "created_by": "System"
        }
        role_id = self.mongo.insert_one("roles", role)
        logger.debug(f"Created role: id={role_id}, name=system_admin, permissions=[{permission_id}]")
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
            "password_hash": password_hash,
            "status": 1,  # ACTIVE
            "email_verified": True,
            "roles": [{
                "role_id": role_id,
                "tenant_id": tenant_id,
                "assigned_at": datetime.now(),
                "assigned_by": "System"
            }],
            "created_at": datetime.now(),
            "created_by": "System"
        }
        user_id = self.mongo.insert_one("users", user)
        logger.debug(f"Created user: id={user_id}, email={TestConfig.DEFAULT_ADMIN_EMAIL}, roles=[{role_id}]")
        return user_id
