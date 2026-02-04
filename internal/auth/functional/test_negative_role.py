"""
Functional tests for Role Service - Negative test cases (rbac.proto - RoleService).
Tests error scenarios: duplicate names, cross-tenant access, system role protection.
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
from auth.v1 import rbac_pb2, rbac_pb2_grpc, role_pb2, tenant_pb2, tenant_pb2_grpc, user_pb2_grpc, user_pb2
from infra.v1 import infra_pb2
from logger import get_logger
from db.mongo_client import MongoDBClient
from db.redis_client import RedisClient
from seeders.system_seeder import SystemSeeder
from helpers.db_injection import inject_role, inject_tenant, inject_user

# Test logger
logger = get_logger("tests.role.negative")


@pytest.mark.role
@pytest.mark.negative
@pytest.mark.integration
class TestRoleManagementErrors:
    """Test role management error scenarios."""

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
            self.role_id = system_data["role_id"]

    def test_create_role_duplicate_name(self):
        """Test CreateRole with role name that already exists in tenant."""
        logger.info("Step 1: Injecting first role directly into MongoDB")

        # Pre-test: Inject first role directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            inject_role(
                mongo,
                tenant_id=self.tenant_id,
                name="DuplicateRole",
                permissions=["*:*"],
                description="First role",
                type=2,  # ROLE_TYPE_CUSTOM
                status=1,  # ROLE_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to create second role with same name")
        # Act - Test CreateRole gRPC endpoint with duplicate name
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            logger.info("Step 2: Attempting to create second role with same name")
            role2 = role_pb2.Role(
                tenant_id=self.tenant_id,
                name="DuplicateRole",  # Same name
                description="Second role",
                type=role_pb2.ROLE_TYPE_CUSTOM,
                permissions=["*:*"],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request2 = rbac_pb2.CreateRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role=role2
            )

            logger.info("Step 3: Expecting ALREADY_EXISTS error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateRole(create_request2)

            assert exc_info.value.code() == grpc.StatusCode.ALREADY_EXISTS
            logger.info("Step 4: Test completed - received expected ALREADY_EXISTS error")

    def test_create_role_invalid_tenant(self):
        """Test CreateRole with non-existent tenant_id."""
        logger.info("Step 1: Attempting to create role with invalid tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            role = role_pb2.Role(
                tenant_id="000000000000000000000000",  # Non-existent tenant
                name="TestRole",
                description="Test role",
                type=role_pb2.ROLE_TYPE_CUSTOM,
                permissions=["*:*"],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = rbac_pb2.CreateRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role=role
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateRole(create_request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_create_role_empty_name(self):
        """Test CreateRole with missing role name."""
        logger.info("Step 1: Attempting to create role with empty name")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            role = role_pb2.Role(
                tenant_id=self.tenant_id,
                name="",  # Empty name
                description="Test role",
                type=role_pb2.ROLE_TYPE_CUSTOM,
                permissions=["*:*"],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = rbac_pb2.CreateRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role=role
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateRole(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_role_invalid_permission(self):
        """Test CreateRole with a permission string not in the registry."""
        logger.info("Step 1: Attempting to create role with invalid permission string")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            role = role_pb2.Role(
                tenant_id=self.tenant_id,
                name="TestRole",
                description="Test role",
                type=role_pb2.ROLE_TYPE_CUSTOM,
                permissions=["nonexistent:garbage"],  # Not in the permission registry
                status=role_pb2.ROLE_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = rbac_pb2.CreateRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role=role
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateRole(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_role_cross_tenant_access(self):
        """Test CreateRole fails when a user from another tenant tries to create a role in our tenant."""
        logger.info("Step 1: Injecting another tenant and user directly into MongoDB")

        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant for Role",
                slug="other-tenant-role",
                status=1,  # TENANT_STATUS_ACTIVE
                contact={"email": "test@erp.com"},
                created_by=self.admin_user_id
            )

            other_tenant_user_id = inject_user(
                mongo,
                tenant_id=other_tenant_id,
                email="otherrbac@example.com",
                username="otherrbacuser",
                password="vK9!xQp#2A@ZLr8",
                status=1,  # USER_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to create role in our tenant as the other-tenant user")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            role_stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            role = role_pb2.Role(
                tenant_id=self.tenant_id,  # Our tenant
                name="CrossTenantRole",
                description="Test role",
                type=role_pb2.ROLE_TYPE_CUSTOM,
                permissions=["*:*"],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = rbac_pb2.CreateRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=other_tenant_id, user_id=other_tenant_user_id),
                role=role
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                role_stub.CreateRole(create_request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_get_role_nonexistent(self):
        """Test GetRole with invalid role_id."""
        logger.info("Step 1: Attempting to get non-existent role")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            request = rbac_pb2.GetRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role_id="000000000000000000000000",  # Non-existent role
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetRole(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_role_cross_tenant_access(self):
        """Test GetRole for role from different tenant."""
        logger.info("Step 1: Injecting another tenant with role directly into MongoDB")

        # Pre-test: Inject tenant and role directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Inject other tenant
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant for GetRole",
                slug="other-tenant-getrole",
                status=1,  # TENANT_STATUS_ACTIVE
                contact={"email": "test@erp.com"},
                created_by=self.admin_user_id
            )

            other_tenant_user_id = inject_user(
                mongo,
                tenant_id=other_tenant_id,
                email="otherrbac@example.com",
                username="otherrbacuser",
                password="vK9!xQp#2A@ZLr8",
                status=1,  # USER_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to get role from different tenant")
        # Act - Test GetRole gRPC endpoint with cross-tenant access
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            role_stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            request = rbac_pb2.GetRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=other_tenant_id, user_id=other_tenant_user_id),
                role_id=self.role_id,
                target_tenant_id=self.tenant_id  # Different tenant
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                role_stub.GetRole(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_list_roles_invalid_pagination(self):
        """Test ListRoles with invalid page_token."""
        logger.info("Step 1: Attempting to list roles with invalid pagination")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            request = rbac_pb2.ListRolesRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                target_tenant_id=self.tenant_id
                # Note: pagination with invalid page_token would go here
            )

            logger.info("Step 2: Note - Skipping test as pagination validation may not be implemented yet")
            pytest.skip("Pagination validation may not be implemented in ListRoles endpoint yet")

    def test_update_role_nonexistent(self):
        """Test UpdateRole with invalid role_id."""
        logger.info("Step 1: Attempting to update non-existent role")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            role = role_pb2.Role(
                id="000000000000000000000000",  # Non-existent role
                tenant_id=self.tenant_id,
                name="UpdatedRole",
                description="Updated role",
                type=role_pb2.ROLE_TYPE_CUSTOM,
                permissions=["*:*"],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            request = rbac_pb2.UpdateRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role=role
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdateRole(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_update_role_duplicate_name(self):
        """Test UpdateRole changing to existing role name."""
        logger.info("Step 1: Injecting two roles directly into MongoDB")

        # Pre-test: Inject two roles directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Inject role1
            inject_role(
                mongo,
                tenant_id=self.tenant_id,
                name="RoleOne",
                permissions=["*:*"],
                description="First role",
                type=2,  # ROLE_TYPE_CUSTOM
                status=1,  # ROLE_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

            # Inject role2
            role2_id = inject_role(
                mongo,
                tenant_id=self.tenant_id,
                name="RoleTwo",
                permissions=["*:*"],
                description="Second role",
                type=2,  # ROLE_TYPE_CUSTOM
                status=1,  # ROLE_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to update role2 name to role1's name")
        # Act - Test UpdateRole gRPC endpoint with duplicate name
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            # Try to update role2's name to role1's name
            role2 = role_pb2.Role(
                id=role2_id,
                tenant_id=self.tenant_id,
                name="RoleOne",  # Duplicate name
                description="Second role",
                type=role_pb2.ROLE_TYPE_CUSTOM,
                permissions=["*:*"],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            update_request = rbac_pb2.UpdateRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role=role2
            )

            logger.info("Step 3: Expecting ALREADY_EXISTS error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdateRole(update_request)

            assert exc_info.value.code() == grpc.StatusCode.ALREADY_EXISTS
            logger.info("Step 4: Test completed - received expected ALREADY_EXISTS error")

    def test_update_role_cross_tenant_access(self):
        """Test UpdateRole for role from different tenant."""
        logger.info("Step 1: Injecting another tenant with role directly into MongoDB")

        # Pre-test: Inject tenant and role directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Inject other tenant
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant for UpdateRole",
                slug="other-tenant-updaterole",
                status=1,  # TENANT_STATUS_ACTIVE
                contact={"email": "test@erp.com"},
                created_by=self.admin_user_id
            )

            other_tenant_user_id = inject_user(
                mongo,
                tenant_id=other_tenant_id,
                email="otherrbac@example.com",
                username="otherrbacuser",
                password="vK9!xQp#2A@ZLr8",
                status=1,  # USER_STATUS_ACTIVE
                created_by=self.admin_user_id
            )


        logger.info("Step 2: Attempting to update role from different tenant")
        # Act - Test UpdateRole gRPC endpoint with cross-tenant access
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            role_stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            updated_role = role_pb2.Role(
                id=self.role_id,
                tenant_id=self.tenant_id,
                name="UpdatedCrossTenantRole",
                description="Updated role",
                type=role_pb2.ROLE_TYPE_CUSTOM,
                permissions=[],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                created_by=other_tenant_user_id
            )

            request = rbac_pb2.UpdateRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=other_tenant_id, user_id=other_tenant_user_id),
                role=updated_role
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                role_stub.UpdateRole(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_delete_role_nonexistent(self):
        """Test DeleteRole with invalid role_id."""
        logger.info("Step 1: Attempting to delete non-existent role")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            request = rbac_pb2.DeleteRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role_id="000000000000000000000000",  # Non-existent role
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.DeleteRole(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_delete_system_role(self):
        """Test DeleteRole attempting to delete system_admin role."""
        logger.info("Step 1: Attempting to delete system admin role")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            # Use the system role ID from setup
            request = rbac_pb2.DeleteRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role_id=self.role_id,  # System admin role
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.DeleteRole(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 3: Test completed - received expected PERMISSION_DENIED error")

    def test_delete_role_with_assigned_users(self):
        """Test DeleteRole with role assigned to active users."""
        logger.info("Step 1: Injecting role and user directly into MongoDB")

        # Pre-test: Inject role and user with that role directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Inject role
            role_id = inject_role(
                mongo,
                tenant_id=self.tenant_id,
                name="AssignedRole",
                permissions=["*:*"],
                description="Role with assigned users",
                type=2,  # ROLE_TYPE_CUSTOM
                status=1,  # ROLE_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

            # Inject user with this role
            inject_user(
                mongo,
                tenant_id=self.tenant_id,
                email="userrole@example.com",
                username="userrole",
                password="vK9!xQp#2A@ZLr8",
                profile={
                    "first_name": "User",
                    "last_name": "Role",
                    "display_name": "User Role"
                },
                roles=[{
                    "tenant_id": self.tenant_id,
                    "role_id": role_id,
                    "assigned_at": datetime.now(),
                    "assigned_by": self.admin_user_id
                }],
                status=1,  # USER_STATUS_ACTIVE
                preferences={
                    "language": "en",
                    "timezone": "UTC",
                    "theme": "light",
                    "notifications": {"email": True, "push": False, "sms": False}
                },
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to delete role with assigned users")
        # Act - Test DeleteRole gRPC endpoint with role assigned to users
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            role_stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            logger.info("Step 2: Attempting to delete role with assigned users")
            delete_request = rbac_pb2.DeleteRoleRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                role_id=role_id,
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                role_stub.DeleteRole(delete_request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")
