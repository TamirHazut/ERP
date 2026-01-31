"""
Functional tests for Permission Service - Negative test cases (rbac.proto - PermissionService).
Tests error scenarios: duplicate permission strings, invalid formats, system permission protection.
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

from grpc_client import GrpcClient
from bson import ObjectId
from datetime import datetime
from config import TestConfig
from auth.v1 import rbac_pb2, rbac_pb2_grpc, permission_pb2, tenant_pb2, tenant_pb2_grpc
from infra.v1 import infra_pb2
from logger import get_logger
from db.mongo_client import MongoDBClient
from db.redis_client import RedisClient
from seeders.system_seeder import SystemSeeder
from helpers.db_injection import inject_permission, inject_tenant

# Test logger
logger = get_logger("tests.permission.negative")


@pytest.mark.permission
@pytest.mark.negative
@pytest.mark.integration
class TestPermissionManagementErrors:
    """Test permission management error scenarios."""

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
            self.permission_id = system_data["permission_id"]

    def test_create_permission_duplicate_string(self):
        """Test CreatePermission with permission string that already exists in tenant."""
        logger.info("Step 1: Injecting first permission directly into MongoDB")

        # Pre-test: Inject first permission directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            inject_permission(
                mongo,
                tenant_id=self.tenant_id,
                permission_string="users:duplicate",
                resource="users",
                action="duplicate",
                display_name="Duplicate Permission",
                description="First permission",
                status=1,  # PERMISSION_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to create second permission with same string")
        # Act - Test CreatePermission gRPC endpoint with duplicate permission string
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            logger.info("Step 2: Attempting to create second permission with same string")
            permission2 = permission_pb2.Permission(
                tenant_id=self.tenant_id,
                display_name="Another Duplicate Permission",
                permission_string="users:duplicate",  # Same permission string
                description="Second permission",
                resource="users",
                action="duplicate",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request2 = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission=permission2
            )

            logger.info("Step 3: Expecting ALREADY_EXISTS error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreatePermission(create_request2)

            assert exc_info.value.code() == grpc.StatusCode.ALREADY_EXISTS
            logger.info("Step 4: Test completed - received expected ALREADY_EXISTS error")

    def test_create_permission_invalid_tenant(self):
        """Test CreatePermission with non-existent tenant_id."""
        logger.info("Step 1: Attempting to create permission with invalid tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            permission = permission_pb2.Permission(
                tenant_id="000000000000000000000000",  # Non-existent tenant
                display_name="Test Permission",
                permission_string="users:test",
                description="Test permission",
                resource="users",
                action="test",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission=permission
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreatePermission(create_request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_create_permission_empty_string(self):
        """Test CreatePermission with missing permission string."""
        logger.info("Step 1: Attempting to create permission with empty permission string")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            permission = permission_pb2.Permission(
                tenant_id=self.tenant_id,
                display_name="Empty Permission",
                permission_string="",  # Empty permission string
                description="Test permission",
                resource="users",
                action="test",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission=permission
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreatePermission(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_permission_invalid_format(self):
        """Test CreatePermission with invalid permission string format (no colon)."""
        logger.info("Step 1: Attempting to create permission with invalid format")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            permission = permission_pb2.Permission(
                tenant_id=self.tenant_id,
                display_name="Invalid Format Permission",
                permission_string="invalid-format-no-colon",  # No colon separator
                description="Test permission",
                resource="users",
                action="test",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission=permission
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreatePermission(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_permission_empty_resource(self):
        """Test CreatePermission with permission string \":action\" (empty resource)."""
        logger.info("Step 1: Attempting to create permission with empty resource")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            permission = permission_pb2.Permission(
                tenant_id=self.tenant_id,
                display_name="Empty Resource Permission",
                permission_string=":action",  # Empty resource
                description="Test permission",
                resource="",
                action="action",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission=permission
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreatePermission(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_permission_empty_action(self):
        """Test CreatePermission with permission string \"resource:\" (empty action)."""
        logger.info("Step 1: Attempting to create permission with empty action")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            permission = permission_pb2.Permission(
                tenant_id=self.tenant_id,
                display_name="Empty Action Permission",
                permission_string="resource:",  # Empty action
                description="Test permission",
                resource="resource",
                action="",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission=permission
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreatePermission(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_get_permission_nonexistent(self):
        """Test GetPermission with invalid permission_id."""
        logger.info("Step 1: Attempting to get non-existent permission")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission_id="000000000000000000000000",  # Non-existent permission
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetPermission(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_permission_cross_tenant_access(self):
        """Test GetPermission for permission from different tenant."""
        logger.info("Step 1: Injecting another tenant with permission directly into MongoDB")

        # Pre-test: Inject tenant and permission directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Inject other tenant
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant for Permission",
                slug="other-tenant-permission",
                status=1,  # TENANT_STATUS_ACTIVE
                contact={"email": "test@erp.com"},
                created_by=self.admin_user_id
            )

            # Inject permission for other tenant
            other_tenant_permission_id = inject_permission(
                mongo,
                tenant_id=other_tenant_id,
                permission_string="test:read",
                resource="test",
                action="read",
                display_name="Test Read",
                description="Test permission for other tenant",
                status=1,  # PERMISSION_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to get permission from different tenant")
        # Act - Test GetPermission gRPC endpoint with cross-tenant access
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            permission_stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission_id=other_tenant_permission_id,
                target_tenant_id=other_tenant_id  # Different tenant
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                permission_stub.GetPermission(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_list_permissions_invalid_pagination(self):
        """Test ListPermissions with page_size > max_limit."""
        logger.info("Step 1: Attempting to list permissions with invalid pagination")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            request = rbac_pb2.ListPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                target_tenant_id=self.tenant_id
                # Note: pagination with excessive page_size would go here
            )

            logger.info("Step 2: Note - Skipping test as pagination validation may not be implemented yet")
            pytest.skip("Pagination validation may not be implemented in ListPermissions endpoint yet")

    def test_update_permission_nonexistent(self):
        """Test UpdatePermission with invalid permission_id."""
        logger.info("Step 1: Attempting to update non-existent permission")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            permission = permission_pb2.Permission(
                id="000000000000000000000000",  # Non-existent permission
                tenant_id=self.tenant_id,
                display_name="Updated Permission",
                permission_string="users:updated",
                description="Updated permission",
                resource="users",
                action="updated",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            request = rbac_pb2.UpdatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission=permission
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdatePermission(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_update_permission_duplicate_string(self):
        """Test UpdatePermission changing to existing permission string."""
        logger.info("Step 1: Injecting two permissions directly into MongoDB")

        # Pre-test: Inject two permissions directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Inject permission1
            inject_permission(
                mongo,
                tenant_id=self.tenant_id,
                permission_string="users:one",
                resource="users",
                action="one",
                display_name="Permission One",
                description="First permission",
                status=1,  # PERMISSION_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

            # Inject permission2
            permission2_id = inject_permission(
                mongo,
                tenant_id=self.tenant_id,
                permission_string="users:two",
                resource="users",
                action="two",
                display_name="Permission Two",
                description="Second permission",
                status=1,  # PERMISSION_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to update permission2 string to permission1's string")
        # Act - Test UpdatePermission gRPC endpoint with duplicate permission string
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            # Try to update permission2's string to permission1's string
            permission2 = permission_pb2.Permission(
                id=permission2_id,
                tenant_id=self.tenant_id,
                display_name="Permission Two",
                permission_string="users:one",  # Duplicate string
                description="Second permission",
                resource="users",
                action="one",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            update_request = rbac_pb2.UpdatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission=permission2
            )

            logger.info("Step 3: Expecting ALREADY_EXISTS error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdatePermission(update_request)

            assert exc_info.value.code() == grpc.StatusCode.ALREADY_EXISTS
            logger.info("Step 4: Test completed - received expected ALREADY_EXISTS error")

    def test_update_permission_cross_tenant_access(self):
        """Test UpdatePermission for permission from different tenant."""
        logger.info("Step 1: Injecting another tenant with permission directly into MongoDB")

        # Pre-test: Inject tenant and permission directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Inject other tenant
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant for UpdatePermission",
                slug="other-tenant-updateperm",
                status=1,  # TENANT_STATUS_ACTIVE
                contact={"email": "test@erp.com"},
                created_by=self.admin_user_id
            )

            # Inject permission for other tenant
            other_tenant_permission_id = inject_permission(
                mongo,
                tenant_id=other_tenant_id,
                permission_string="test:update",
                resource="test",
                action="update",
                display_name="Test Update",
                description="Test permission for other tenant",
                status=1,  # PERMISSION_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to update permission from different tenant")
        # Act - Test UpdatePermission gRPC endpoint with cross-tenant access
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            permission_stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            updated_permission = permission_pb2.Permission(
                id=other_tenant_permission_id,
                tenant_id=other_tenant_id,
                display_name="Updated Cross Tenant Permission",
                permission_string="test:update",
                description="Updated permission",
                resource="test",
                action="update",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            request = rbac_pb2.UpdatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission=updated_permission
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                permission_stub.UpdatePermission(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_delete_permission_nonexistent(self):
        """Test DeletePermission with invalid permission_id."""
        logger.info("Step 1: Attempting to delete non-existent permission")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            request = rbac_pb2.DeletePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission_id="000000000000000000000000",  # Non-existent permission
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.DeletePermission(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_delete_system_permission(self):
        """Test DeletePermission attempting to delete *:* permission."""
        logger.info("Step 1: Attempting to delete system wildcard permission")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            # Use the system permission ID from setup (wildcard permission)
            request = rbac_pb2.DeletePermissionRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                permission_id=self.permission_id,  # System wildcard permission
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.DeletePermission(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 3: Test completed - received expected PERMISSION_DENIED error")
