"""
Functional tests for Permission Service - Negative test cases (rbac.proto - PermissionService).
Permissions are code-defined in the registry — only GetPermission and ListPermissions RPCs exist.
Tests error scenarios: invalid permission strings, cross-tenant access denied.
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
# Add auth functional path for helpers
auth_functional_path = os.path.dirname(__file__)
sys.path.insert(0, auth_functional_path)

from lib.functional.grpc_client import GrpcClient
from lib.functional.config import TestConfig
from lib.model.auth.v1 import rbac_pb2, rbac_pb2_grpc
from lib.model.infra.v1 import infra_pb2
from lib.functional.logger import get_logger
from lib.functional.db.mongo_client import MongoDBClient
from lib.functional.db.redis_client import RedisClient
from lib.functional.seeders.system_seeder import SystemSeeder
from helpers.db_injection import inject_role, inject_tenant, inject_user

# Test logger
logger = get_logger("tests.permission.negative")


@pytest.mark.permission
@pytest.mark.negative
@pytest.mark.integration
class TestPermissionManagementErrors:
    """Test permission read error scenarios against the code-defined registry."""

    @pytest.fixture(autouse=True)
    def setup(self, clean_database):
        """Setup test data before each test."""

        database = os.getenv("AUTH_DB_NAME", "auth_db_test")

        # Seed system data
        with MongoDBClient(database) as mongo, RedisClient() as redis:
            seeder = SystemSeeder(mongo, redis)
            system_data = seeder.seed_all()

            self.tenant_id = system_data["tenant_id"]
            self.admin_user_id = system_data["user_id"]

    def test_get_permission_nonexistent(self):
        """Test GetPermission with a permission string not in the registry."""
        logger.info("Step 1: Attempting to get non-existent permission string")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission_id="nonexistent:garbage",  # Not in the registry
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetPermission(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_permission_invalid_format(self):
        """Test GetPermission with a malformed permission string (no colon)."""
        logger.info("Step 1: Attempting to get permission with invalid format")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission_id="invalid-no-colon",
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetPermission(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_permission_token_resource_rejected(self):
        """Test GetPermission with 'token:read' — token is not a valid resource type."""
        logger.info("Step 1: Attempting to get permission for removed token resource")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission_id="token:read",
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetPermission(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_permission_cross_tenant_access(self):
        """Test GetPermission fails when a user from a new tenant tries to read a permission targeting the default tenant."""
        logger.info("Step 1: Injecting new tenant and new user directly into MongoDB")

        # Pre-test: Inject a new tenant and a user belonging to that tenant
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant for Permission",
                slug="other-tenant-permission",
                status=1,  # TENANT_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

            other_user_id = inject_user(
                mongo,
                tenant_id=other_tenant_id,
                email="other@erp.com",
                username="otheruser",
                status=1,  # USER_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to get permission targeting default tenant using the new user")
        # Act - The new user (other tenant) tries to read a permission targeting the default tenant
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            permission_stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=other_tenant_id, user_id=other_user_id),
                permission_id="user:read",  # Valid registry permission
                target_tenant_id=self.tenant_id  # Default tenant — cross-tenant
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                permission_stub.GetPermission(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_delete_protected_role_blocked(self):
        """Test DeleteRole on a protected role is blocked for non-system tenant."""
        logger.info("Step 1: Injecting a protected role and a non-system tenant user")

        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Create a second tenant with a user
            other_tenant_id = inject_tenant(
                mongo,
                name="Tenant For Protected Role Test",
                slug="tenant-protected-role",
                status=1,
                created_by=self.admin_user_id
            )

            # Inject a protected role into that tenant
            protected_role_id = inject_role(
                mongo,
                tenant_id=other_tenant_id,
                name="protected_role",
                permissions=[],
                protected=True,
                created_by=self.admin_user_id
            )

            # Inject a user in that tenant
            other_user_id = inject_user(
                mongo,
                tenant_id=other_tenant_id,
                email="roletest@erp.com",
                username="roletestuser",
                status=1,
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to delete protected role as non-system tenant user")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            request = rbac_pb2.DeleteRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=other_tenant_id, user_id=other_user_id),
                role_id=protected_role_id,
                target_tenant_id=other_tenant_id
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.DeleteRole(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")
