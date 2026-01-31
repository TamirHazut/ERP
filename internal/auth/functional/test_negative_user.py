"""
Functional tests for User Service - Negative test cases (user.proto).
Tests error scenarios: duplicate emails, invalid tenants, permission violations, tenant isolation.
"""
from bson import ObjectId
from datetime import datetime
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
from config import TestConfig
from auth.v1 import user_pb2, user_pb2_grpc, tenant_pb2, tenant_pb2_grpc
from infra.v1 import infra_pb2
from logger import get_logger
from db.mongo_client import MongoDBClient
from db.redis_client import RedisClient
from seeders.system_seeder import SystemSeeder
from helpers.db_injection import inject_user, inject_tenant, inject_role

# Test logger
logger = get_logger("tests.user.negative")


@pytest.mark.user
@pytest.mark.negative
@pytest.mark.integration
class TestUserManagementErrors:
    """Test user management error scenarios."""

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

    def test_create_user_duplicate_email(self):
        """Test CreateUser with email that already exists in tenant."""
        logger.info("Step 1: Injecting first user directly into MongoDB")

        # Pre-test: Inject user directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            inject_user(
                mongo,
                tenant_id=self.tenant_id,
                email="duplicate@example.com",
                username="user1",
                password_hash="hashed_password",
                profile={"first_name": "User", "last_name": "One"},
                status=1,  # USER_STATUS_ACTIVE
                preferences={
                    "language": "en", "timezone": "UTC", "theme": "light",
                    "notifications": {"email": True, "push": False, "sms": False}
                },
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to create second user with same email")
        # Act - Test CreateUser gRPC endpoint with duplicate email
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user2 = user_pb2.User(
                tenant_id=self.tenant_id,
                email="duplicate@example.com",  # Same email
                username="user2",
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="User", last_name="Two"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            create_request2 = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user2
            )

            logger.info("Step 3: Expecting ALREADY_EXISTS error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateUser(create_request2)

            assert exc_info.value.code() == grpc.StatusCode.ALREADY_EXISTS
            logger.info("Step 4: Test completed - received expected ALREADY_EXISTS error")

    def test_create_user_duplicate_username(self):
        """Test CreateUser with username that already exists in tenant."""
        logger.info("Step 1: Injecting first user directly into MongoDB")

        # Pre-test: Inject user directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            inject_user(
                mongo,
                tenant_id=self.tenant_id,
                email="user1@example.com",
                username="duplicateuser",
                password_hash="hashed_password",
                profile={"first_name": "User", "last_name": "One"},
                status=1,  # USER_STATUS_ACTIVE
                preferences={
                    "language": "en", "timezone": "UTC", "theme": "light",
                    "notifications": {"email": True, "push": False, "sms": False}
                },
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to create second user with same username")
        # Act - Test CreateUser gRPC endpoint with duplicate username
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user2 = user_pb2.User(
                tenant_id=self.tenant_id,
                email="user2@example.com",
                username="duplicateuser",  # Same username
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="User", last_name="Two"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            create_request2 = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user2
            )

            logger.info("Step 3: Expecting ALREADY_EXISTS error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateUser(create_request2)

            assert exc_info.value.code() == grpc.StatusCode.ALREADY_EXISTS
            logger.info("Step 4: Test completed - received expected ALREADY_EXISTS error")

    def test_create_user_invalid_tenant(self):
        """Test CreateUser with non-existent tenant_id."""
        logger.info("Step 1: Attempting to create user with invalid tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user = user_pb2.User(
                tenant_id="000000000000000000000000",  # Non-existent tenant
                email="user@example.com",
                username="testuser",
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="Test", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            create_request = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateUser(create_request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_create_user_missing_email(self):
        """Test CreateUser with empty email field."""
        logger.info("Step 1: Attempting to create user with empty email")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user = user_pb2.User(
                tenant_id=self.tenant_id,
                email="",  # Empty email
                username="testuser",
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="Test", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            create_request = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateUser(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_user_missing_password(self):
        """Test CreateUser with empty password field."""
        logger.info("Step 1: Attempting to create user with empty password")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user = user_pb2.User(
                tenant_id=self.tenant_id,
                email="user@example.com",
                username="testuser",
                password_hash="",  # Empty password
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="Test", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            create_request = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateUser(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_user_invalid_email_format(self):
        """Test CreateUser with malformed email (no @)."""
        logger.info("Step 1: Attempting to create user with malformed email")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user = user_pb2.User(
                tenant_id=self.tenant_id,
                email="invalid-email-no-at-sign",  # Invalid email format
                username="testuser",
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="Test", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            create_request = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateUser(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_user_weak_password(self):
        """Test CreateUser with password too short (<8 chars)."""
        logger.info("Step 1: Attempting to create user with weak password")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user = user_pb2.User(
                tenant_id=self.tenant_id,
                email="user@example.com",
                username="testuser",
                password_hash="short",  # Too short password
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="Test", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            create_request = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateUser(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_user_invalid_role(self):
        """Test CreateUser with non-existent role."""
        logger.info("Step 1: Attempting to create user with non-existent role")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user = user_pb2.User(
                tenant_id=self.tenant_id,
                email="user@example.com",
                username="testuser",
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                roles=[user_pb2.UserRole(
                    tenant_id=self.tenant_id,
                    role_id="000000000000000000000000",  # Non-existent role
                    assigned_by=self.admin_user_id
                )],
                profile=user_pb2.UserProfile(first_name="Test", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            create_request = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateUser(create_request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_create_user_cross_tenant_role(self):
        """Test CreateUser with role from different tenant."""
        logger.info("Step 1: Injecting second tenant with role directly into MongoDB")

        # Pre-test: Inject another tenant and role directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant",
                slug="other-tenant",
                status=1,  # TENANT_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

            # Inject a role for the other tenant
            other_tenant_role_id = inject_role(
                mongo,
                tenant_id=other_tenant_id,
                name="OtherTenantRole",
                permissions=[],
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to create user with role from different tenant")
        # Act - Test CreateUser gRPC endpoint with cross-tenant role
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            user_stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user = user_pb2.User(
                tenant_id=self.tenant_id,  # Our tenant
                email="user@example.com",
                username="testuser",
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                roles=[user_pb2.UserRole(
                    tenant_id=other_tenant_id,  # Different tenant's role
                    role_id=other_tenant_role_id,
                    assigned_by=self.admin_user_id
                )],
                profile=user_pb2.UserProfile(first_name="Test", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            create_request = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                user_stub.CreateUser(create_request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_get_user_nonexistent(self):
        """Test GetUser with invalid user_id."""
        logger.info("Step 1: Attempting to get non-existent user")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            request = user_pb2.GetUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                target_tenant_id=self.tenant_id,
                account_id="000000000000000000000000"  # Non-existent user
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetUser(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_user_cross_tenant_access(self):
        """Test GetUser for user from different tenant."""
        logger.info("Step 1: Creating user in different tenant")

        # Pre-test: Create another tenant with user
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            tenant_stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            new_tenant = tenant_pb2.Tenant(
                name="Other Tenant 2",
                slug="other-tenant-2",
                status=tenant_pb2.TENANT_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                contact=tenant_pb2.ContactInfo(email="test@erp.com")
            )

            create_tenant_request = tenant_pb2.CreateTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant=new_tenant
            )

            create_tenant_response = tenant_stub.CreateTenant(create_tenant_request)
            other_tenant_id = create_tenant_response.tenant_id

            # Get user from other tenant
            database = os.getenv("AUTH_DB_NAME", "auth_db_test")
            with MongoDBClient(database) as mongo:
                users_collection = mongo.get_collection("users")
                other_tenant_user = users_collection.find_one({"tenant_id": other_tenant_id})
                other_tenant_user_id = other_tenant_user["_id"]

            logger.info("Step 2: Attempting to get user from different tenant")
            user_stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            request = user_pb2.GetUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                target_tenant_id=other_tenant_id,  # Different tenant
                account_id=other_tenant_user_id
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                user_stub.GetUser(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_get_user_missing_id(self):
        """Test GetUser with empty user_id."""
        logger.info("Step 1: Attempting to get user with empty user_id")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            request = user_pb2.GetUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                target_tenant_id=self.tenant_id,
                account_id=""  # Empty user_id
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetUser(request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_list_users_invalid_pagination(self):
        """Test ListUsers with invalid pagination (page_size = -1)."""
        logger.info("Step 1: Attempting to list users with invalid pagination")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            request = user_pb2.ListUsersRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                target_tenant_id=self.tenant_id
                # Note: pagination parameter would go here if proto supported it
            )

            logger.info("Step 2: Note - Skipping test as pagination validation may not be implemented yet")
            pytest.skip("Pagination validation may not be implemented in ListUsers endpoint yet")

    def test_list_users_cross_tenant_filter(self):
        """Test ListUsers filtering by different tenant."""
        logger.info("Step 1: Injecting another tenant directly into MongoDB")

        # Pre-test: Inject another tenant directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant 3",
                slug="other-tenant-3",
                status=1,  # TENANT_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to list users from different tenant")
        # Act - Test ListUsers gRPC endpoint with cross-tenant access
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            user_stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            request = user_pb2.ListUsersRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                target_tenant_id=other_tenant_id  # Different tenant
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                user_stub.ListUsers(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_update_user_nonexistent(self):
        """Test UpdateUser with invalid user_id."""
        logger.info("Step 1: Attempting to update non-existent user")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user = user_pb2.User(
                id="000000000000000000000000",  # Non-existent user
                tenant_id=self.tenant_id,
                email="user@example.com",
                username="testuser",
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="Test", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            request = user_pb2.UpdateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdateUser(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_update_user_duplicate_email(self):
        """Test UpdateUser changing to existing email."""
        logger.info("Step 1: Injecting two users directly into MongoDB")

        # Pre-test: Inject two users directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Inject user1
            inject_user(
                mongo,
                tenant_id=self.tenant_id,
                email="user1@example.com",
                username="user1",
                password_hash="hashed_password",
                profile={"first_name": "User", "last_name": "One"},
                status=1,  # USER_STATUS_ACTIVE
                preferences={
                    "language": "en", "timezone": "UTC", "theme": "light",
                    "notifications": {"email": True, "push": False, "sms": False}
                },
                created_by=self.admin_user_id
            )

            # Inject user2
            user2_id = inject_user(
                mongo,
                tenant_id=self.tenant_id,
                email="user2@example.com",
                username="user2",
                password_hash="hashed_password",
                profile={"first_name": "User", "last_name": "Two"},
                status=1,  # USER_STATUS_ACTIVE
                preferences={
                    "language": "en", "timezone": "UTC", "theme": "light",
                    "notifications": {"email": True, "push": False, "sms": False}
                },
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to update user2 email to user1's email")
        # Act - Test UpdateUser gRPC endpoint with duplicate email
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user2 = user_pb2.User(
                id=user2_id,
                tenant_id=self.tenant_id,
                email="user1@example.com",  # Duplicate email
                username="user2",
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="User", last_name="Two"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            update_request = user_pb2.UpdateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user2
            )

            logger.info("Step 3: Expecting ALREADY_EXISTS error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdateUser(update_request)

            assert exc_info.value.code() == grpc.StatusCode.ALREADY_EXISTS
            logger.info("Step 4: Test completed - received expected ALREADY_EXISTS error")

    def test_update_user_cross_tenant_access(self):
        """Test UpdateUser for user from different tenant."""
        logger.info("Step 1: Injecting another tenant with user directly into MongoDB")

        # Pre-test: Inject another tenant and user directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant 4",
                slug="other-tenant-4",
                status=1,  # TENANT_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

            other_user_id = inject_user(
                mongo,
                tenant_id=other_tenant_id,
                email="other@example.com",
                username="otheruser4",
                password_hash="hashed_password",
                status=1,  # USER_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to update user from different tenant")
        # Act - Test UpdateUser gRPC endpoint with cross-tenant access
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            user_stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            updated_user = user_pb2.User(
                id=other_user_id,
                tenant_id=other_tenant_id,
                email="other@example.com",
                username="otheruser4",
                password_hash="hashed_password",
                status=user_pb2.USER_STATUS_SUSPENDED,  # Try to suspend
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="Updated", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            request = user_pb2.UpdateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=updated_user
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                user_stub.UpdateUser(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

    def test_update_user_invalid_status(self):
        """Test UpdateUser with invalid status value."""
        logger.info("Step 1: Injecting user directly into MongoDB")

        # Pre-test: Inject user directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            user_id = inject_user(
                mongo,
                tenant_id=self.tenant_id,
                email="user@example.com",
                username="testuser",
                password_hash="hashed_password",
                profile={"first_name": "Test", "last_name": "User"},
                status=1,  # USER_STATUS_ACTIVE
                preferences={
                    "language": "en", "timezone": "UTC", "theme": "light",
                    "notifications": {"email": True, "push": False, "sms": False}
                },
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to update user with invalid status")
        # Act - Test UpdateUser gRPC endpoint with invalid status
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            user = user_pb2.User(
                id=user_id,
                tenant_id=self.tenant_id,
                email="user@example.com",
                username="testuser",
                password_hash="hashed_password",
                status=999,  # Invalid status
                created_by=self.admin_user_id,
                profile=user_pb2.UserProfile(first_name="Test", last_name="User"),
                preferences=user_pb2.UserPreferences(
                    language="en", timezone="UTC", theme="light",
                    notifications=user_pb2.NotificationSettings(email=True, push=False, sms=False)
                )
            )

            update_request = user_pb2.UpdateUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                user=user
            )

            logger.info("Step 3: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdateUser(update_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 4: Test completed - received expected INVALID_ARGUMENT error")

    def test_delete_user_nonexistent(self):
        """Test DeleteUser with invalid user_id."""
        logger.info("Step 1: Attempting to delete non-existent user")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            request = user_pb2.DeleteUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                target_tenant_id=self.tenant_id,
                account_id="000000000000000000000000"  # Non-existent user
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.DeleteUser(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_delete_user_cross_tenant_access(self):
        """Test DeleteUser for user from different tenant."""
        logger.info("Step 1: Injecting another tenant with user directly into MongoDB")

        # Pre-test: Inject another tenant and user directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            other_tenant_id = inject_tenant(
                mongo,
                name="Other Tenant 5",
                slug="other-tenant-5",
                status=1,  # TENANT_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

            other_tenant_user_id = inject_user(
                mongo,
                tenant_id=other_tenant_id,
                email="other5@example.com",
                username="otheruser5",
                password_hash="hashed_password",
                status=1,  # USER_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to delete user from different tenant")
        # Act - Test DeleteUser gRPC endpoint with cross-tenant access
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            user_stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            request = user_pb2.DeleteUserRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                target_tenant_id=other_tenant_id,  # Different tenant
                account_id=other_tenant_user_id
            )

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                user_stub.DeleteUser(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")
