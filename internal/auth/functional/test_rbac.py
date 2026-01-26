"""
Functional tests for VerificationService (rbac.proto).
Tests: CheckPermissions, HasPermission, GetUserPermissions, GetUserRoles, IsSystemTenantUser.
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

# Test logger
logger = get_logger("tests.rbac")


@pytest.mark.auth
@pytest.mark.integration
class TestRBACVerification:
    """Test RBAC verification flows (happy path)."""

    @pytest.fixture(autouse=True)
    def setup(self, clean_database):
        """Setup test data before each test."""
        from db.mongo_client import MongoDBClient
        from seeders.system_seeder import SystemSeeder

        database = os.getenv("AUTH_DB_NAME", "auth_db_test")

        # Seed system data (admin user has system_admin role with "*:*" permission)
        with MongoDBClient(database) as mongo:
            seeder = SystemSeeder(mongo)
            system_data = seeder.seed_all()

            self.tenant_id = system_data["tenant_id"]
            self.user_id = system_data["user_id"]
            self.role_id = system_data["role_id"]
            self.permission_id = system_data["permission_id"]

        self.user_email = TestConfig.DEFAULT_ADMIN_EMAIL
        self.user_password = TestConfig.DEFAULT_ADMIN_PASSWORD

    def test_check_permissions_success(self):
        """Test checking multiple permissions at once."""
        logger.info("Step 1: Pre-test - admin user has *:* permission (all access)")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            logger.info("Step 2: Act - calling CheckPermissions RPC")
            # Act: Check multiple permissions
            check_request = rbac_pb2.CheckPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permissions=["products:read", "products:write", "orders:read"]
            )
            check_response = stub.CheckPermissions(check_request)

            logger.info("Step 3: Assert - validating all permissions are granted")
            # Assert: Admin with *:* has all permissions
            assert check_response.permissions["products:read"] is True
            assert check_response.permissions["products:write"] is True
            assert check_response.permissions["orders:read"] is True
            logger.info("All permissions granted (admin has *:* permission)")

            logger.info("Step 4: CheckPermissions test completed successfully")

    def test_has_permission_success(self):
        """Test checking a single permission."""
        logger.info("Step 1: Pre-test - admin user has *:* permission (all access)")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            logger.info("Step 2: Act - calling HasPermission RPC")
            # Act: Check single permission
            has_request = rbac_pb2.HasPermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                permission="products:read",
                target_tenant_id=self.tenant_id
            )
            has_response = stub.HasPermission(has_request)

            logger.info("Step 3: Assert - validating permission is granted")
            # Assert
            assert has_response.has_permission is True
            logger.info("Permission granted (admin has *:* permission)")

            logger.info("Step 4: HasPermission test completed successfully")

    def test_get_user_permissions_success(self):
        """Test getting all permissions for a user."""
        logger.info("Step 1: Pre-test - admin user has system_admin role with *:* permission")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            logger.info("Step 2: Act - calling GetUserPermissions RPC")
            # Act: Get user permissions
            get_perms_request = rbac_pb2.GetUserPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                )
            )
            get_perms_response = stub.GetUserPermissions(get_perms_request)

            logger.info("Step 3: Assert - validating user has *:* permission")
            # Assert: Should contain *:* permission
            assert "*:*" in get_perms_response.permissions
            assert get_perms_response.permissions["*:*"] is True
            logger.info(f"User has {len(get_perms_response.permissions)} permissions, including *:*")

            logger.info("Step 4: GetUserPermissions test completed successfully")

    def test_get_user_roles_success(self):
        """Test getting all roles assigned to a user."""
        logger.info("Step 1: Pre-test - admin user has system_admin role from seeder")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            logger.info("Step 2: Act - calling GetUserRoles RPC")
            # Act: Get user roles
            get_roles_request = rbac_pb2.GetUserRolesRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                )
            )
            get_roles_response = stub.GetUserRoles(get_roles_request)

            logger.info("Step 3: Assert - validating user has system_admin role")
            # Assert: Should contain system_admin role
            assert len(get_roles_response.role_ids) >= 1
            assert self.role_id in get_roles_response.role_ids
            logger.info(f"User has {len(get_roles_response.role_ids)} role(s), including system_admin")

            logger.info("Step 4: GetUserRoles test completed successfully")

    def test_is_system_tenant_user_success(self):
        """Test checking if a tenant is the system tenant."""
        logger.info("Step 1: Pre-test - SystemSeeder creates tenant with name 'System Tenant'")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            logger.info("Step 2: Act - calling IsSystemTenantUser RPC")
            # Act: Check if tenant is system tenant
            is_system_request = rbac_pb2.IsSystemTenantUserRequest(
                tenant_id=self.tenant_id
            )
            is_system_response = stub.IsSystemTenantUser(is_system_request)

            logger.info("Step 3: Assert - validating tenant is system tenant")
            # Assert: Should be system tenant
            assert is_system_response.is_system_tenant is True
            logger.info("Tenant is system tenant (as expected)")

            logger.info("Step 4: IsSystemTenantUser test completed successfully")
