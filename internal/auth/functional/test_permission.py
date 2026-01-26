"""
Functional tests for PermissionService (rbac.proto).
Tests: CreatePermission, UpdatePermission, GetPermission, ListPermissions, DeletePermission.
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
from auth.v1 import rbac_pb2, rbac_pb2_grpc, permission_pb2
from infra.v1 import infra_pb2
from logger import get_logger

# Test logger
logger = get_logger("tests.permission")


@pytest.mark.auth
@pytest.mark.integration
class TestPermissionManagement:
    """Test permission management flows (happy path)."""

    @pytest.fixture(autouse=True)
    def setup(self, clean_database):
        """Setup test data before each test."""
        from db.mongo_client import MongoDBClient
        from seeders.system_seeder import SystemSeeder

        database = os.getenv("AUTH_DB_NAME", "auth_db_test")

        with MongoDBClient(database) as mongo:
            seeder = SystemSeeder(mongo)
            system_data = seeder.seed_all()

            self.tenant_id = system_data["tenant_id"]
            self.user_id = system_data["user_id"]
            self.role_id = system_data["role_id"]
            self.permission_id = system_data["permission_id"]

        self.user_email = TestConfig.DEFAULT_ADMIN_EMAIL
        self.user_password = TestConfig.DEFAULT_ADMIN_PASSWORD

    def test_create_permission_success(self):
        """Test creating a new permission."""
        logger.info("Step 1: Pre-test - preparing permission data")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            logger.info("Step 2: Act - calling CreatePermission RPC")
            # Act: Create a new permission
            permission = permission_pb2.Permission(
                tenant_id=self.tenant_id,
                resource="products",
                action="read",
                permission_string="products:read",
                display_name="Read Products",
                description="View product list",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                is_dangerous=False
            )

            create_request = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission=permission
            )

            create_response = stub.CreatePermission(create_request)

            logger.info("Step 3: Assert - validating permission was created")
            # Assert
            assert create_response.permission_id != ""
            assert len(create_response.permission_id) > 0
            logger.info(f"Created permission with ID: {create_response.permission_id}")

            logger.info("Step 4: CreatePermission test completed successfully")

    def test_update_permission_success(self):
        """Test updating an existing permission."""
        logger.info("Step 1: Pre-test - creating a permission to update")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            # Pre-test: Create a permission
            permission = permission_pb2.Permission(
                tenant_id=self.tenant_id,
                resource="products",
                action="write",
                permission_string="products:write",
                display_name="Write Products",
                description="Create and edit products",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                is_dangerous=False
            )

            create_request = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission=permission
            )
            create_response = stub.CreatePermission(create_request)
            permission_id = create_response.permission_id

            logger.info("Step 2: Act - calling UpdatePermission RPC with modified data")
            # Act: Update the permission
            updated_permission = permission_pb2.Permission(
                id=permission_id,
                tenant_id=self.tenant_id,
                resource="products",
                action="write",
                permission_string="products:write",
                display_name="Write Products",
                description="Updated: View product catalog",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                is_dangerous=False
            )

            update_request = rbac_pb2.UpdatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission=updated_permission
            )
            update_response = stub.UpdatePermission(update_request)

            logger.info("Step 3: Assert - validating update was successful")
            # Assert
            assert update_response.success is True

            logger.info("Step 4: Verify - checking description was updated")
            # Verify: Get the permission and check the description
            get_request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission_id=permission_id,
                target_tenant_id=self.tenant_id
            )
            get_response = stub.GetPermission(get_request)
            assert get_response.description == "Updated: View product catalog"
            logger.info("Description successfully updated")

            logger.info("Step 5: UpdatePermission test completed successfully")

    def test_get_permission_success(self):
        """Test retrieving a permission by ID."""
        logger.info("Step 1: Pre-test - creating a permission to retrieve")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            # Pre-test: Create a permission
            permission = permission_pb2.Permission(
                tenant_id=self.tenant_id,
                resource="orders",
                action="create",
                permission_string="orders:create",
                display_name="Create Orders",
                description="Create new orders",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                is_dangerous=False
            )

            create_request = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission=permission
            )
            create_response = stub.CreatePermission(create_request)
            permission_id = create_response.permission_id

            logger.info("Step 2: Act - calling GetPermission RPC")
            # Act: Get the permission
            get_request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission_id=permission_id,
                target_tenant_id=self.tenant_id
            )
            get_response = stub.GetPermission(get_request)

            logger.info("Step 3: Assert - validating retrieved permission data")
            # Assert
            assert get_response.id == permission_id
            assert get_response.resource == "orders"
            assert get_response.action == "create"
            assert get_response.permission_string == "orders:create"
            logger.info(f"Retrieved permission: {get_response.permission_string}")

            logger.info("Step 4: GetPermission test completed successfully")

    def test_list_permissions_success(self):
        """Test listing all permissions for a tenant."""
        logger.info("Step 1: Pre-test - creating 3 test permissions")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            # Pre-test: Create 3 permissions
            permissions_data = [
                ("products", "read", "products:read"),
                ("products", "write", "products:write"),
                ("orders", "delete", "orders:delete")
            ]

            for resource, action, perm_string in permissions_data:
                permission = permission_pb2.Permission(
                    tenant_id=self.tenant_id,
                    resource=resource,
                    action=action,
                    permission_string=perm_string,
                    display_name=f"{action.title()} {resource.title()}",
                    description=f"{action.title()} permission for {resource}",
                    status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                    is_dangerous=False
                )

                create_request = rbac_pb2.CreatePermissionRequest(
                    identifier=infra_pb2.UserIdentifier(
                        tenant_id=self.tenant_id,
                        user_id=self.user_id
                    ),
                    permission=permission
                )
                stub.CreatePermission(create_request)
                logger.info(f"Created permission: {perm_string}")

            logger.info("Step 2: Act - calling ListPermissions RPC")
            # Act: List all permissions
            list_request = rbac_pb2.ListPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                target_tenant_id=self.tenant_id
            )
            list_response = stub.ListPermissions(list_request)

            logger.info("Step 3: Assert - validating permission list")
            # Assert: Should have at least 4 permissions (3 created + 1 system "*:*" from seeder)
            assert len(list_response.permissions) >= 4

            # Verify our created permissions are in the list
            permission_strings = [p.permission_string for p in list_response.permissions]
            for _, _, perm_string in permissions_data:
                assert perm_string in permission_strings
            logger.info(f"Found {len(list_response.permissions)} total permissions, including our 3 test permissions")

            logger.info("Step 4: ListPermissions test completed successfully")

    def test_delete_permission_success(self):
        """Test deleting a permission."""
        logger.info("Step 1: Pre-test - creating a permission to delete")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.PermissionServiceStub(client.get_channel())

            # Pre-test: Create a permission
            permission = permission_pb2.Permission(
                tenant_id=self.tenant_id,
                resource="test_resource",
                action="test_action",
                permission_string="test_resource:test_action",
                display_name="Test Permission",
                description="This permission will be deleted",
                status=permission_pb2.PERMISSION_STATUS_ACTIVE,
                is_dangerous=False
            )

            create_request = rbac_pb2.CreatePermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission=permission
            )
            create_response = stub.CreatePermission(create_request)
            permission_id = create_response.permission_id

            logger.info("Step 2: Act - calling DeletePermission RPC")
            # Act: Delete the permission
            delete_request = rbac_pb2.DeletePermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission_id=permission_id,
                target_tenant_id=self.tenant_id
            )
            delete_response = stub.DeletePermission(delete_request)

            logger.info("Step 3: Assert - validating deletion was successful")
            # Assert
            assert delete_response.success is True

            logger.info("Step 4: Verify - checking permission no longer exists")
            # Verify: Try to get the permission and expect an error
            get_request = rbac_pb2.GetPermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission_id=permission_id,
                target_tenant_id=self.tenant_id
            )
            try:
                stub.GetPermission(get_request)
                # If we get here, the permission still exists (test should fail)
                assert False, "Permission should not exist after deletion"
            except grpc.RpcError as e:
                # Expected: permission not found
                assert e.code() in [grpc.StatusCode.NOT_FOUND, grpc.StatusCode.UNKNOWN]
                logger.info("Permission successfully deleted (not found)")

            logger.info("Step 5: DeletePermission test completed successfully")
