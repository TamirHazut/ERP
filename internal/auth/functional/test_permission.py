"""
Functional tests for PermissionService (rbac.proto).
Permissions are code-defined in a registry — only GetPermission and ListPermissions RPCs exist.
Tests: GetPermission, ListPermissions.
"""
import pytest
import grpc
import sys
import os

# Add infra functional path to sys.path for imports
infra_functional_path = os.path.join(os.path.dirname(__file__), '../../infra/functional')
sys.path.insert(0, infra_functional_path)
# Add proto path for proto imports
sys.path.insert(0, os.path.join(infra_functional_path, 'proto'))

from grpc_client import GrpcClient
from config import TestConfig
from auth.v1 import rbac_pb2, rbac_pb2_grpc
from infra.v1 import infra_pb2
from logger import get_logger
from db.mongo_client import MongoDBClient
from seeders.system_seeder import SystemSeeder
from db.redis_client import RedisClient

# Test logger
logger = get_logger("tests.permission")


@pytest.mark.auth
@pytest.mark.integration
class TestPermissionManagement:
    """Test permission read flows backed by the code-defined registry."""

    @pytest.fixture(autouse=True)
    def setup(self, clean_database):
        """Setup test data before each test."""

        database = os.getenv("AUTH_DB_NAME", "auth_db_test")

        with MongoDBClient(database) as mongo, RedisClient() as redis:
            seeder = SystemSeeder(mongo, redis)
            system_data = seeder.seed_all()

            self.tenant_id = system_data["tenant_id"]
            self.user_id = system_data["user_id"]

        self.user_email = TestConfig.DEFAULT_ADMIN_EMAIL
        self.user_password = TestConfig.DEFAULT_ADMIN_PASSWORD

    def test_get_permission_success(self):
        """Test retrieving a known registry permission by its permission string."""
        logger.info("Step 1: Act - calling GetPermission with permission_id='user:read'")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            get_request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission_id="user:read",
                target_tenant_id=self.tenant_id
            )
            get_response = stub.GetPermission(get_request)

            logger.info("Step 2: Assert - validating retrieved permission fields")
            assert get_response.id == "user:read"
            assert get_response.resource == "user"
            assert get_response.action == "read"
            assert get_response.permission_string == "user:read"
            assert get_response.display_name == "User Read"
            assert get_response.protected is True
            assert get_response.is_dangerous is False
            logger.info(f"Retrieved permission: {get_response.permission_string}")

            logger.info("Step 3: GetPermission test completed successfully")

    def test_get_permission_admin_wildcard(self):
        """Test retrieving the *:* admin wildcard permission."""
        logger.info("Step 1: Act - calling GetPermission with permission_id='*:*'")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            get_request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission_id="*:*",
                target_tenant_id=self.tenant_id
            )
            get_response = stub.GetPermission(get_request)

            logger.info("Step 2: Assert - validating admin wildcard fields")
            assert get_response.id == "*:*"
            assert get_response.permission_string == "*:*"
            assert get_response.is_dangerous is True
            assert get_response.protected is True
            logger.info("Retrieved admin wildcard permission: *:*")

            logger.info("Step 3: GetPermission admin wildcard test completed successfully")

    def test_list_permissions_success(self):
        """Test listing permissions returns the full code-defined registry."""
        logger.info("Step 1: Act - calling ListPermissions")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            list_request = rbac_pb2.ListPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                target_tenant_id=self.tenant_id
            )
            list_response = stub.ListPermissions(list_request)

            logger.info("Step 2: Assert - validating permission list contains known registry entries")
            permission_strings = [p.permission_string for p in list_response.permissions]

            # Core registry permissions that must always be present
            expected_permissions = [
                "*:*",
                "user:create", "user:read", "user:update", "user:delete",
                "role:create", "role:read", "role:update", "role:delete",
                "permission:read",
                "tenant:create", "tenant:read", "tenant:update", "tenant:delete",
                "config:read", "config:update",
            ]

            for perm in expected_permissions:
                assert perm in permission_strings, f"Expected '{perm}' in ListPermissions response"

            logger.info(f"Found {len(list_response.permissions)} total permissions, all expected entries present")

            # Verify all returned permissions are marked protected
            for p in list_response.permissions:
                assert p.protected is True, f"Permission '{p.permission_string}' should be protected"

            # Verify *:* is the only dangerous permission
            dangerous = [p.permission_string for p in list_response.permissions if p.is_dangerous]
            assert dangerous == ["*:*"]

            logger.info("Step 3: ListPermissions test completed successfully")
