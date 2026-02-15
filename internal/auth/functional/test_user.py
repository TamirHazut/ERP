"""
Functional tests for User Service (user.proto).
Tests: CreateUser, GetUser, ListUsers, UpdateUser, DeleteUser.
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

from lib.functional.grpc_client import GrpcClient
from lib.functional.config import TestConfig
from lib.model.auth.v1 import user_pb2, user_pb2_grpc
from lib.model.infra.v1 import infra_pb2
from lib.functional.logger import get_logger
from lib.functional.db.mongo_client import MongoDBClient
from lib.functional.db.redis_client import RedisClient
from lib.functional.seeders.system_seeder import SystemSeeder
from helpers.db_injection import inject_user

# Test logger
logger = get_logger("tests.user")


@pytest.mark.user
@pytest.mark.integration
class TestUserManagement:
    """Test user management operations (happy path)."""

    @pytest.fixture(autouse=True)
    def setup(self, clean_database):
        """Setup test data before each test."""

        database = os.getenv("AUTH_DB_NAME", "auth_db_test")  # Separate test database

        # Seed system data (tenant, permission, role, admin user)
        with MongoDBClient(database) as mongo, RedisClient() as redis:
            seeder = SystemSeeder(mongo, redis)
            system_data = seeder.seed_all()

            # Store IDs for use in tests (these are real MongoDB ObjectIDs)
            self.tenant_id = system_data["tenant_id"]
            self.admin_user_id = system_data["user_id"]
            self.role_id = system_data["role_id"]

    def test_create_user_success(self):
        """Test successful user creation."""
        logger.info("Step 1: Creating user creation request")

        # Arrange
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            # Create user object
            new_user = user_pb2.User(
                tenant_id=self.tenant_id,
                email="testuser@example.com",
                username="testuser",
                password="vK9!xQp#2A@ZLr8",  # In real scenario, this would be hashed
                profile=user_pb2.UserProfile(
                    first_name="Test",
                    last_name="User",
                    display_name="Test User"
                ),
                status=user_pb2.USER_STATUS_ACTIVE,
                email_verified=False,
                phone_verified=False,
                mfa_enabled=False,
                created_by=self.admin_user_id,
                preferences=user_pb2.UserPreferences(
                    language="en",
                    timezone="UTC",
                    theme="light",
                    notifications=user_pb2.NotificationSettings(
                        email=True,
                        push=False,
                        sms=False
                    )
                )
            )

            logger.info("Step 2: Sending CreateUser RPC request")
            # Act
            request = user_pb2.CreateUserRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                user=new_user
            )

            response = stub.CreateUser(request)

            logger.info("Step 3: Validating response - checking user_id received")
            # Assert
            assert response.user_id is not None
            assert response.user_id != ""

            logger.info("Step 4: Verifying user exists in MongoDB")
            # Post-test verification
            database = os.getenv("AUTH_DB_NAME", "auth_db_test")
            with MongoDBClient(database) as mongo:
                users_collection = mongo.get_collection("users")
                user_doc = users_collection.find_one({"_id": ObjectId(response.user_id)})

                assert user_doc is not None
                assert user_doc["email"] == "testuser@example.com"
                assert user_doc["username"] == "testuser"
                assert user_doc["tenant_id"] == self.tenant_id
                assert user_doc["status"] == user_pb2.USER_STATUS_ACTIVE

            logger.info("Step 5: CreateUser test completed successfully")

    def test_get_user_success(self):
        """Test successful user retrieval."""
        logger.info("Step 1: Injecting user directly into MongoDB")

        # Pre-test: Inject user directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            created_user_id = inject_user(
                mongo,
                tenant_id=self.tenant_id,
                email="getuser@example.com",
                username="getuser",
                password="vK9!xQp#2A@ZLr8",
                profile={
                    "first_name": "Get",
                    "last_name": "User",
                    "display_name": "Get User"
                },
                status=1,  # USER_STATUS_ACTIVE
                preferences={
                    "language": "en",
                    "timezone": "UTC",
                    "theme": "light",
                    "notifications": {"email": True, "push": False, "sms": False}
                },
                created_by=self.admin_user_id
            )

        logger.info(f"Step 2: Retrieving user with ID: {created_user_id}")
        # Act - Test GetUser gRPC endpoint
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            get_request = user_pb2.GetUserRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                target_tenant_id=self.tenant_id,
                account_id=created_user_id
            )

            user_response = stub.GetUser(get_request)

            logger.info("Step 3: Validating retrieved user fields")
            # Assert
            assert user_response.id == created_user_id
            assert user_response.email == "getuser@example.com"
            assert user_response.username == "getuser"
            assert user_response.tenant_id == self.tenant_id
            assert user_response.status == user_pb2.USER_STATUS_ACTIVE
            assert user_response.profile.first_name == "Get"
            assert user_response.profile.last_name == "User"

            logger.info("Step 4: GetUser test completed successfully")

    def test_list_users_success(self):
        """Test successful user listing with pagination."""
        logger.info("Step 1: Injecting multiple users directly into MongoDB")

        # Pre-test: Inject 3 users directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            for i in range(1, 4):
                inject_user(
                    mongo,
                    tenant_id=self.tenant_id,
                    email=f"listuser{i}@example.com",
                    username=f"listuser{i}",
                    password="vK9!xQp#2A@ZLr8",
                    profile={
                        "first_name": f"List{i}",
                        "last_name": "User",
                        "display_name": f"List User {i}"
                    },
                    status=1,  # USER_STATUS_ACTIVE
                    preferences={
                        "language": "en",
                        "timezone": "UTC",
                        "theme": "light",
                        "notifications": {"email": True, "push": False, "sms": False}
                    },
                    created_by=self.admin_user_id
                )

        logger.info("Step 2: Listing all users for tenant")
        # Act - Test ListUsers gRPC endpoint
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            list_request = user_pb2.ListUsersRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                target_tenant_id=self.tenant_id
            )

            list_response = stub.ListUsers(list_request)

            logger.info("Step 3: Validating user list count")
            # Assert - should have 4 users (1 admin + 3 created)
            assert len(list_response.users) >= 3

            # Verify our created users are in the list
            user_emails = [user.email for user in list_response.users]
            assert "listuser1@example.com" in user_emails
            assert "listuser2@example.com" in user_emails
            assert "listuser3@example.com" in user_emails

            logger.info("Step 4: ListUsers test completed successfully")

    def test_update_user_success(self):
        """Test successful user update."""
        logger.info("Step 1: Injecting user directly into MongoDB")

        # Pre-test: Inject user directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            created_user_id = inject_user(
                mongo,
                tenant_id=self.tenant_id,
                email="updateuser@example.com",
                username="updateuser",
                password="vK9!xQp#2A@ZLr8",
                profile={
                    "first_name": "Update",
                    "last_name": "User",
                    "display_name": "Update User"
                },
                status=1,  # USER_STATUS_ACTIVE
                preferences={
                    "language": "en",
                    "timezone": "UTC",
                    "theme": "light",
                    "notifications": {"email": True, "push": False, "sms": False}
                },
                created_by=self.admin_user_id
            )

        logger.info(f"Step 2: Updating user with ID: {created_user_id}")
        # Act - Test UpdateUser gRPC endpoint
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            updated_user = user_pb2.User(
                id=created_user_id,
                tenant_id=self.tenant_id,
                email="updateuser@example.com",
                username="updateuser",
                password="vK9!xQp#2A@ZLr8",
                profile=user_pb2.UserProfile(
                    first_name="Updated",
                    last_name="User",
                    display_name="Updated User"
                ),
                status=user_pb2.USER_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                preferences=user_pb2.UserPreferences(
                    language="fr",  # Changed from "en"
                    timezone="Europe/Paris",  # Changed from "UTC"
                    theme="dark",  # Changed from "light"
                    notifications=user_pb2.NotificationSettings(
                        email=True,
                        push=True,  # Changed from False
                        sms=False
                    )
                )
            )

            update_request = user_pb2.UpdateUserRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                user=updated_user
            )

            update_response = stub.UpdateUser(update_request)

            logger.info("Step 3: Validating update response")
            assert update_response.updated is True

            logger.info("Step 4: Retrieving updated user to verify changes")
            # Post-test verification - get the user and verify changes
            get_request = user_pb2.GetUserRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                target_tenant_id=self.tenant_id,
                account_id=created_user_id
            )

            user_response = stub.GetUser(get_request)

            logger.info("Step 5: Verifying updated fields")
            assert user_response.profile.first_name == "Updated"
            assert user_response.preferences.language == "fr"
            assert user_response.preferences.timezone == "Europe/Paris"
            assert user_response.preferences.theme == "dark"
            assert user_response.preferences.notifications.push is True

            logger.info("Step 6: UpdateUser test completed successfully")

    def test_delete_user_success(self):
        """Test successful user deletion (soft delete - status change)."""
        logger.info("Step 1: Injecting user directly into MongoDB")

        # Pre-test: Inject user directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            created_user_id = inject_user(
                mongo,
                tenant_id=self.tenant_id,
                email="deleteuser@example.com",
                username="deleteuser",
                password="vK9!xQp#2A@ZLr8",
                profile={
                    "first_name": "Delete",
                    "last_name": "User",
                    "display_name": "Delete User"
                },
                status=1,  # USER_STATUS_ACTIVE
                preferences={
                    "language": "en",
                    "timezone": "UTC",
                    "theme": "light",
                    "notifications": {"email": True, "push": False, "sms": False}
                },
                created_by=self.admin_user_id
            )

        logger.info(f"Step 2: Deleting user with ID: {created_user_id}")
        # Act - Test DeleteUser gRPC endpoint
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = user_pb2_grpc.UserServiceStub(client.get_channel())

            delete_request = user_pb2.DeleteUserRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                target_tenant_id=self.tenant_id,
                account_id=created_user_id
            )

            delete_response = stub.DeleteUser(delete_request)

            logger.info("Step 3: Validating delete response")
            assert delete_response.deleted is True

            logger.info("Step 4: Verifying user deleted from MongoDB")
            # Post-test verification - check user deleted from MongoDB
            with MongoDBClient(database) as mongo:
                users_collection = mongo.get_collection("users")
                user_doc = users_collection.find_one({"_id": ObjectId(created_user_id)})

                # User should be deleted
                assert user_doc is None

            logger.info("Step 5: DeleteUser test completed successfully")
