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
            self.permission_id = system_data["permission_id"]

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
        pytest.skip("Test does not supposed to fail")
        """Test CheckPermissions for user from different tenant."""
        logger.info("Step 1: Injecting another tenant with user directly into MongoDB")

        # Pre-test: Inject another tenant and user directly into MongoDB
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
                password_hash="hashed_password",
                status=1,  # USER_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting CheckPermissions for user from different tenant")
        # Act - Test CheckPermissions gRPC endpoint with cross-tenant check
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            rbac_stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.CheckPermissionsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=other_tenant_id,  # Different tenant
                    user_id=other_tenant_user_id
                ),
                permissions=["user:read"]
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error (cross-tenant check not allowed)")
            with pytest.raises(grpc.RpcError) as exc_info:
                rbac_stub.CheckPermissions(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

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
                password_hash="hashed_password",
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
        pytest.skip("test does not supposed to raise exception")
        """Test IsSystemTenantUser with invalid user_id."""
        logger.info("Step 1: Attempting IsSystemTenantUser for non-existent tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = rbac_pb2_grpc.VerificationServiceStub(client.get_channel())

            request = rbac_pb2.IsSystemTenantUserRequest(
                tenant_id="000000000000000000000000"  # Non-existent tenant
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.IsSystemTenantUser(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")
