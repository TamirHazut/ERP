"""
Functional tests for RBAC Verification Service - Negative test cases (rbac.proto).
Tests error scenarios: invalid users, deleted entities, cross-tenant verification.
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
from auth.v1 import rbac_pb2, rbac_pb2_grpc, user_pb2, tenant_pb2, tenant_pb2_grpc
from infra.v1 import infra_pb2
from logger import get_logger
from db.mongo_client import MongoDBClient
from db.redis_client import RedisClient
from seeders.system_seeder import SystemSeeder
from helpers.db_injection import inject_tenant, inject_user

# Test logger
logger = get_logger("tests.rbac.negative")


@pytest.mark.rbac
@pytest.mark.negative
@pytest.mark.integration
class TestRBACVerificationErrors:
    """Test RBAC verification error scenarios."""

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

    def test_check_permissions_nonexistent_user(self):
        """Test CheckPermissions with invalid user_id."""
        logger.info("Step 1: Attempting CheckPermissions for non-existent user")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.CheckPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id="000000000000000000000000"  # Non-existent user
                ),
                permissions=["user:read"]
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CheckPermissions(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_check_permissions_inactive_user(self):
        """Test CheckPermissions with user status=INACTIVE."""
        logger.info("Step 1: Setting user status to INACTIVE")

        # Pre-test: Set user to inactive
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            users_collection = mongo.get_collection("users")
            users_collection.update_one(
                {"_id": ObjectId(self.admin_user_id)},
                {"$set": {"status": user_pb2.USER_STATUS_INACTIVE}}
            )

        logger.info("Step 2: Attempting CheckPermissions for inactive user")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.CheckPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                permissions=["user:read"]
            )

            logger.info("Step 3: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CheckPermissions(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 4: Test completed - received expected UNAUTHENTICATED error")

    def test_check_permissions_suspended_user(self):
        """Test CheckPermissions with user status=SUSPENDED."""
        logger.info("Step 1: Setting user status to SUSPENDED")

        # Pre-test: Set user to suspended
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            users_collection = mongo.get_collection("users")
            users_collection.update_one(
                {"_id": ObjectId(self.admin_user_id)},
                {"$set": {"status": user_pb2.USER_STATUS_SUSPENDED}}
            )

        logger.info("Step 2: Attempting CheckPermissions for suspended user")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.CheckPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                permissions=["user:read"]
            )

            logger.info("Step 3: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CheckPermissions(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 4: Test completed - received expected UNAUTHENTICATED error")

    def test_check_permissions_deleted_user(self):
        """Test CheckPermissions with user status=DELETED."""
        logger.info("Step 1: Deleting user from database")

        # Pre-test: Delete user
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            users_collection = mongo.get_collection("users")
            users_collection.delete_one({"_id": ObjectId(self.admin_user_id)})

        logger.info("Step 2: Attempting CheckPermissions for deleted user")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.CheckPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                permissions=["user:read"]
            )

            logger.info("Step 3: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CheckPermissions(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 4: Test completed - received expected NOT_FOUND error")

    def test_check_permissions_invalid_tenant(self):
        """Test CheckPermissions with non-existent tenant_id."""
        logger.info("Step 1: Attempting CheckPermissions with invalid tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.CheckPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id="000000000000000000000000",  # Non-existent tenant
                    user_id=self.admin_user_id
                ),
                permissions=["user:read"]
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CheckPermissions(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_check_permissions_cross_tenant_check(self):
        """Test CheckPermissions for user with no roles returns denied in response map.

        CheckPermissions has no target_tenant_id parameter — it always checks permissions
        within the caller's own tenant. A user with no roles gets a success response with
        all requested permissions set to False.
        """
        logger.info("Step 1: Injecting another tenant with user (no roles) directly into MongoDB")

        # Pre-test: Inject another tenant and user with no roles directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant for RBAC",
                slug="other-tenant-rbac",
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

        logger.info("Step 2: Calling CheckPermissions for user with no roles")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            rbac_stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.CheckPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=other_tenant_id,
                    user_id=other_tenant_user_id
                ),
                permissions=["user:read"]
            )

            # CheckPermissions returns a permission map, not a gRPC error, for denied permissions
            response = rbac_stub.CheckPermissions(request)

            logger.info("Step 3: Verifying permission is denied in response map")
            assert response.permissions["user:read"] == False
            logger.info("Step 4: Test completed - user:read correctly reported as denied")

    def test_has_permission_nonexistent_user(self):
        """Test HasPermission with invalid user_id."""
        logger.info("Step 1: Attempting HasPermission for non-existent user")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.HasPermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id="000000000000000000000000"  # Non-existent user
                ),
                permission="user:read",
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.HasPermission(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_has_permission_invalid_permission_format(self):
        """Test HasPermission with malformed permission string."""
        logger.info("Step 1: Attempting HasPermission with invalid permission format")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.HasPermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                permission="invalid-permission-without-colon",  # Invalid format
                target_tenant_id=self.tenant_id
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.HasPermission(request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_has_permission_cross_tenant_check(self):
        """Test HasPermission for user from different tenant."""
        logger.info("Step 1: Injecting another tenant with user directly into MongoDB")

        # Pre-test: Inject another tenant and user directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant for HasPermission",
                slug="other-tenant-hasperm",
                status=1,  # TENANT_STATUS_ACTIVE
                contact={"email": "test@erp.com"},
                created_by=self.admin_user_id
            )

            other_tenant_user_id = inject_user(
                mongo,
                tenant_id=other_tenant_id,
                email="otherhasperm@example.com",
                username="otherhaspermuser",
                password="vK9!xQp#2A@ZLr8",
                status=1,  # USER_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting HasPermission for user from different tenant")
        # Act - Test HasPermission gRPC endpoint with cross-tenant check
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            rbac_stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.HasPermissionRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=other_tenant_id,  # Different tenant
                    user_id=other_tenant_user_id
                ),
                permission="user:read",
                target_tenant_id=other_tenant_id
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                rbac_stub.HasPermission(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_get_user_permissions_nonexistent_user(self):
        """Test GetUserPermissions with invalid user_id."""
        logger.info("Step 1: Attempting GetUserPermissions for non-existent user")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.GetUserPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id="000000000000000000000000"  # Non-existent user
                )
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetUserPermissions(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_user_permissions_deleted_user(self):
        """Test GetUserPermissions with user status=DELETED."""
        logger.info("Step 1: Deleting user from database")

        # Pre-test: Delete user
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            users_collection = mongo.get_collection("users")
            users_collection.delete_one({"_id": ObjectId(self.admin_user_id)})

        logger.info("Step 2: Attempting GetUserPermissions for deleted user")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.GetUserPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                )
            )

            logger.info("Step 3: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetUserPermissions(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 4: Test completed - received expected NOT_FOUND error")

    def test_get_user_roles_nonexistent_user(self):
        """Test GetUserRoles with invalid user_id."""
        logger.info("Step 1: Attempting GetUserRoles for non-existent user")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.GetUserRolesRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id="000000000000000000000000"  # Non-existent user
                )
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetUserRoles(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_user_roles_deleted_user(self):
        """Test GetUserRoles with user status=DELETED."""
        logger.info("Step 1: Deleting user from database")

        # Pre-test: Delete user
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            users_collection = mongo.get_collection("users")
            users_collection.delete_one({"_id": ObjectId(self.admin_user_id)})

        logger.info("Step 2: Attempting GetUserRoles for deleted user")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.GetUserRolesRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                )
            )

            logger.info("Step 3: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetUserRoles(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 4: Test completed - received expected NOT_FOUND error")

    def test_is_system_tenant_user_invalid_user(self):
        """Test IsSystemTenantUser with non-system tenant_id returns false.

        IsSystemTenantUser is a pure boolean comparator: it checks whether the given
        tenant_id matches the system tenant stored in Redis. It does not look up the
        tenant in MongoDB, so non-existent tenant IDs simply return false.
        """
        logger.info("Step 1: Calling IsSystemTenantUser with non-existent tenant_id")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.IsSystemTenantUserRequest(
                tenant_id="000000000000000000000000"  # Non-existent, and not the system tenant
            )

            response = stub.IsSystemTenantUser(request)

            logger.info("Step 2: Verifying non-system tenant is reported as false")
            assert response.is_system_tenant == False
            logger.info("Step 3: Test completed - non-system tenant correctly identified")
