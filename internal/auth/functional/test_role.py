"""
Functional tests for RoleService (rbac.proto).
Tests: CreateRole, UpdateRole, GetRole, ListRoles, DeleteRole.
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
from auth.v1 import rbac_pb2, rbac_pb2_grpc, role_pb2
from infra.v1 import infra_pb2
from logger import get_logger

# Test logger
logger = get_logger("tests.role")


@pytest.mark.auth
@pytest.mark.integration
class TestRoleManagement:
    """Test role management flows (happy path)."""

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

    def test_create_role_success(self):
        """Test creating a new role."""
        logger.info("Step 1: Pre-test - preparing role data")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            logger.info("Step 2: Act - calling CreateRole RPC")
            # Act: Create a new role
            role = role_pb2.Role(
                tenant_id=self.tenant_id,
                name="test_manager",
                description="Test manager role",
                permissions=[],  # Empty permissions for now
                status=role_pb2.ROLE_STATUS_ACTIVE,
                type=role_pb2.ROLE_TYPE_CUSTOM
            )

            create_request = rbac_pb2.CreateRoleRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                role=role
            )

            create_response = stub.CreateRole(create_request)

            logger.info("Step 3: Assert - validating role was created")
            # Assert
            assert create_response.role_id != ""
            assert len(create_response.role_id) > 0
            logger.info(f"Created role with ID: {create_response.role_id}")

            logger.info("Step 4: CreateRole test completed successfully")

    def test_update_role_success(self):
        """Test updating an existing role."""
        logger.info("Step 1: Pre-test - creating a role to update")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            # Pre-test: Create a role
            role = role_pb2.Role(
                tenant_id=self.tenant_id,
                name="test_role",
                description="Test description",
                permissions=[],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                type=role_pb2.ROLE_TYPE_CUSTOM
            )

            create_request = rbac_pb2.CreateRoleRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                role=role
            )
            create_response = stub.CreateRole(create_request)
            role_id = create_response.role_id

            logger.info("Step 2: Act - calling UpdateRole RPC with modified data")
            # Act: Update the role
            updated_role = role_pb2.Role(
                id=role_id,
                tenant_id=self.tenant_id,
                name="test_role",
                description="Updated description",
                permissions=[],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                type=role_pb2.ROLE_TYPE_CUSTOM
            )

            update_request = rbac_pb2.UpdateRoleRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                role=updated_role
            )
            update_response = stub.UpdateRole(update_request)

            logger.info("Step 3: Assert - validating update was successful")
            # Assert
            assert update_response.success is True

            logger.info("Step 4: Verify - checking description was updated")
            # Verify: Get the role and check the description
            get_request = rbac_pb2.GetRoleRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                role_id=role_id,
                target_tenant_id=self.tenant_id
            )
            get_response = stub.GetRole(get_request)
            assert get_response.description == "Updated description"
            logger.info("Description successfully updated")

            logger.info("Step 5: UpdateRole test completed successfully")

    def test_get_role_success(self):
        """Test retrieving a role by ID."""
        logger.info("Step 1: Pre-test - creating a role to retrieve")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            # Pre-test: Create a role
            role = role_pb2.Role(
                tenant_id=self.tenant_id,
                name="test_role",
                description="Test description",
                permissions=[],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                type=role_pb2.ROLE_TYPE_CUSTOM
            )

            create_request = rbac_pb2.CreateRoleRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                role=role
            )
            create_response = stub.CreateRole(create_request)
            role_id = create_response.role_id

            logger.info("Step 2: Act - calling GetRole RPC")
            # Act: Get the role
            get_request = rbac_pb2.GetRoleRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                role_id=role_id,
                target_tenant_id=self.tenant_id
            )
            get_response = stub.GetRole(get_request)

            logger.info("Step 3: Assert - validating retrieved role data")
            # Assert
            assert get_response.id == role_id
            assert get_response.name == "test_role"
            assert get_response.description == "Test description"
            assert get_response.tenant_id == self.tenant_id
            logger.info(f"Retrieved role: {get_response.name}")

            logger.info("Step 4: GetRole test completed successfully")

    def test_list_roles_success(self):
        """Test listing all roles for a tenant."""
        logger.info("Step 1: Pre-test - creating 3 test roles")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            # Pre-test: Create 3 roles
            role_names = ["test_role_1", "test_role_2", "test_role_3"]
            for name in role_names:
                role = role_pb2.Role(
                    tenant_id=self.tenant_id,
                    name=name,
                    description=f"Description for {name}",
                    permissions=[],
                    status=role_pb2.ROLE_STATUS_ACTIVE,
                    type=role_pb2.ROLE_TYPE_CUSTOM
                )

                create_request = rbac_pb2.CreateRoleRequest(
                    identifier=infra_pb2.UserIdentifier(
                        tenant_id=self.tenant_id,
                        user_id=self.user_id
                    ),
                    role=role
                )
                stub.CreateRole(create_request)
                logger.info(f"Created role: {name}")

            logger.info("Step 2: Act - calling ListRoles RPC")
            # Act: List all roles
            list_request = rbac_pb2.ListRolesRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                target_tenant_id=self.tenant_id
            )
            list_response = stub.ListRoles(list_request)

            logger.info("Step 3: Assert - validating role list")
            # Assert: Should have at least 4 roles (3 created + 1 system_admin from seeder)
            assert len(list_response.roles) >= 4

            # Verify our created roles are in the list
            found_roles = [r.name for r in list_response.roles if r.name in role_names]
            assert len(found_roles) == 3
            logger.info(f"Found {len(list_response.roles)} total roles, including our 3 test roles")

            logger.info("Step 4: ListRoles test completed successfully")

    def test_delete_role_success(self):
        """Test deleting a role."""
        logger.info("Step 1: Pre-test - creating a role to delete")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.RoleServiceStub(client.get_channel())

            # Pre-test: Create a role
            role = role_pb2.Role(
                tenant_id=self.tenant_id,
                name="test_role_to_delete",
                description="This role will be deleted",
                permissions=[],
                status=role_pb2.ROLE_STATUS_ACTIVE,
                type=role_pb2.ROLE_TYPE_CUSTOM
            )

            create_request = rbac_pb2.CreateRoleRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                role=role
            )
            create_response = stub.CreateRole(create_request)
            role_id = create_response.role_id

            logger.info("Step 2: Act - calling DeleteRole RPC")
            # Act: Delete the role
            delete_request = rbac_pb2.DeleteRoleRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                role_id=role_id,
                target_tenant_id=self.tenant_id
            )
            delete_response = stub.DeleteRole(delete_request)

            logger.info("Step 3: Assert - validating deletion was successful")
            # Assert
            assert delete_response.success is True

            logger.info("Step 4: Verify - checking role no longer exists")
            # Verify: Try to get the role and expect an error
            get_request = rbac_pb2.GetRoleRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                role_id=role_id,
                target_tenant_id=self.tenant_id
            )
            try:
                stub.GetRole(get_request)
                # If we get here, the role still exists (test should fail)
                assert False, "Role should not exist after deletion"
            except grpc.RpcError as e:
                # Expected: role not found
                assert e.code() in [grpc.StatusCode.NOT_FOUND, grpc.StatusCode.UNKNOWN]
                logger.info("Role successfully deleted (not found)")

            logger.info("Step 5: DeleteRole test completed successfully")
