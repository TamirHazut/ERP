"""
System data seeder for functional tests.
Seeds minimum required data (tenant, admin user, roles, permissions).
"""
from datetime import datetime
from typing import Dict, Any
from ..db.mongo_client import MongoDBClient
from ..config import TestConfig
import bcrypt


class SystemSeeder:
    """Seeds system-level test data."""

    def __init__(self, mongo_client: MongoDBClient):
        self.mongo = mongo_client

    def seed_all(self) -> Dict[str, str]:
        """
        Seed all system data.
        Returns: Dictionary with IDs of created entities.
        """
        # Create tenant
        tenant_id = self.seed_tenant()

        # Create system permission
        permission_id = self.seed_permission(tenant_id)

        # Create system role
        role_id = self.seed_role(tenant_id, permission_id)

        # Create admin user
        user_id = self.seed_admin_user(tenant_id, role_id)

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
            "created_at": datetime.utcnow(),
            "created_by": "System"
        }
        return self.mongo.insert_one("tenants", tenant)

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
            "created_at": datetime.utcnow(),
            "created_by": "System"
        }
        return self.mongo.insert_one("permissions", permission)

    def seed_role(self, tenant_id: str, permission_id: str) -> str:
        """Seed system admin role."""
        role = {
            "tenant_id": tenant_id,
            "name": "system_admin",
            "description": "System administrator role",
            "permissions": [permission_id],
            "status": 1,  # ACTIVE
            "type": 0,  # SYSTEM
            "created_at": datetime.utcnow(),
            "created_by": "System"
        }
        return self.mongo.insert_one("roles", role)

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
                "assigned_at": datetime.utcnow(),
                "assigned_by": "System"
            }],
            "created_at": datetime.utcnow(),
            "created_by": "System"
        }
        return self.mongo.insert_one("users", user)
