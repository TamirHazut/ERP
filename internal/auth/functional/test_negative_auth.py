"""
Functional tests for Auth Service - Negative test cases (auth.proto).
Tests error scenarios: invalid credentials, expired tokens, access violations.
"""
import pytest
import grpc
import sys
import os
import time

# Add infra functional path to sys.path for imports
infra_functional_path = os.path.join(os.path.dirname(__file__), '../../infra/functional')
sys.path.insert(0, infra_functional_path)
# Add proto path for proto imports
sys.path.insert(0, os.path.join(infra_functional_path, 'proto'))

from lib.functional.grpc_client import GrpcClient
from bson import ObjectId
from lib.functional.config import TestConfig
from lib.model.auth.v1 import auth_pb2, auth_pb2_grpc, user_pb2, user_pb2_grpc
from lib.model.infra.v1 import infra_pb2
from lib.functional.logger import get_logger
from lib.functional.db.mongo_client import MongoDBClient
from lib.functional.db.redis_client import RedisClient
from lib.functional.seeders.system_seeder import SystemSeeder

# Test logger
logger = get_logger("tests.auth.negative")


@pytest.mark.auth
@pytest.mark.negative
@pytest.mark.integration
class TestAuthenticationErrors:
    """Test authentication error scenarios."""

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

        self.user_email = TestConfig.DEFAULT_ADMIN_EMAIL
        self.user_password = TestConfig.DEFAULT_ADMIN_PASSWORD

    def test_login_invalid_email(self):
        """Test login with non-existent email."""
        logger.info("Step 1: Attempting login with non-existent email")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email="nonexistent@example.com",
                password=self.user_password
            )

            logger.info("Step 2: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Login(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 3: Test completed - received expected UNAUTHENTICATED error")

    def test_login_invalid_username(self):
        """Test login with non-existent username."""
        logger.info("Step 1: Attempting login with non-existent username")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                username="nonexistentuser",
                password=self.user_password
            )

            logger.info("Step 2: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Login(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 3: Test completed - received expected UNAUTHENTICATED error")

    def test_login_wrong_password(self):
        """Test login with valid user but wrong password."""
        logger.info("Step 1: Attempting login with wrong password")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password="wrong_password_123"
            )

            logger.info("Step 2: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Login(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 3: Test completed - received expected UNAUTHENTICATED error")

    def test_login_inactive_user(self):
        """Test login with INACTIVE user."""
        logger.info("Step 1: Creating inactive user")

        # Pre-test: Create inactive user
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            users_collection = mongo.get_collection("users")
            users_collection.update_one(
                {"_id": ObjectId(self.admin_user_id)},
                {"$set": {"status": user_pb2.USER_STATUS_INACTIVE}}
            )

        logger.info("Step 2: Attempting login with inactive user")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            )

            logger.info("Step 3: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Login(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 4: Test completed - received expected UNAUTHENTICATED error")

    def test_login_suspended_user(self):
        """Test login with SUSPENDED user."""
        logger.info("Step 1: Creating suspended user")

        # Pre-test: Create suspended user
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            users_collection = mongo.get_collection("users")
            users_collection.update_one(
                {"_id": ObjectId(self.admin_user_id)},
                {"$set": {"status": user_pb2.USER_STATUS_SUSPENDED}}
            )

        logger.info("Step 2: Attempting login with suspended user")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            )

            logger.info("Step 3: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Login(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 4: Test completed - received expected UNAUTHENTICATED error")

    def test_login_nonexistent_tenant(self):
        """Test login with invalid tenant_id."""
        logger.info("Step 1: Attempting login with non-existent tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.LoginRequest(
                tenant_id="000000000000000000000000",  # Non-existent tenant ID
                email=self.user_email,
                password=self.user_password
            )

            logger.info("Step 2: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Login(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_login_missing_credentials(self):
        """Test login with empty email/password."""
        logger.info("Step 1: Attempting login with empty credentials")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email="",  # Empty email
                password=""  # Empty password
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Login(request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_logout_expired_token(self):
        """Test logout with expired (invalidated) access token."""
        logger.info("Step 1: Logging in to obtain real tokens")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())
            stub.Login(auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            ))

        logger.info("Step 2: Deleting access token from Redis to simulate expiry")
        with RedisClient() as redis:
            token_key = f"token:{self.tenant_id}:{self.admin_user_id}"
            redis.delete(token_key)

        logger.info("Step 3: Attempting logout — expecting UNAUTHENTICATED")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.LogoutRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                )
            )

            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Logout(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 4: Test completed - received expected UNAUTHENTICATED error")

    def test_refresh_expired_refresh_token(self):
        """Test RefreshToken with expired (invalidated) refresh token."""
        logger.info("Step 1: Logging in to obtain real tokens")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())
            login_response = stub.Login(auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            ))
            refresh_token = login_response.refresh_token.token

        logger.info("Step 2: Deleting refresh token from Redis to simulate expiry")
        with RedisClient() as redis:
            refresh_key = f"refresh_token:{self.tenant_id}:{self.admin_user_id}"
            redis.delete(refresh_key)

        logger.info("Step 3: Attempting token refresh — expecting UNAUTHENTICATED")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.RefreshTokenRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                refresh_token=refresh_token
            )

            with pytest.raises(grpc.RpcError) as exc_info:
                stub.RefreshToken(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 4: Test completed - received expected UNAUTHENTICATED error")

    def test_refresh_invalid_refresh_token(self):
        """Test RefreshToken with malformed refresh token."""
        logger.info("Step 1: Attempting token refresh with malformed refresh token")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.RefreshTokenRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id  # Use real MongoDB ObjectID from setup
                ),
                refresh_token="invalid.refresh.token"
            )

            logger.info("Step 2: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.RefreshToken(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 3: Test completed - received expected UNAUTHENTICATED error")

    def test_verify_invalid_token(self):
        """Test VerifyToken with malformed JWT."""
        logger.info("Step 1: Attempting token verification with malformed JWT")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.VerifyTokenRequest(
                token="invalid.jwt.token"
            )

            logger.info("Step 2: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.VerifyToken(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 3: Test completed - received expected UNAUTHENTICATED error")

    def test_verify_expired_token(self):
        """Test VerifyToken with expired (invalidated) access token."""
        logger.info("Step 1: Logging in to obtain a real access token")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())
            login_response = stub.Login(auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            ))
            access_token = login_response.token.token

        logger.info("Step 2: Deleting access token from Redis to simulate expiry")
        with RedisClient() as redis:
            token_key = f"token:{self.tenant_id}:{self.admin_user_id}"
            redis.delete(token_key)

        logger.info("Step 3: Attempting token verification — expecting UNAUTHENTICATED")
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.VerifyTokenRequest(
                token=access_token
            )

            with pytest.raises(grpc.RpcError) as exc_info:
                stub.VerifyToken(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 4: Test completed - received expected UNAUTHENTICATED error")

    def test_revoke_all_tenant_invalid_tenant(self):
        """Test RevokeAllTenantTokens with non-existent tenant."""
        logger.info("Step 1: Attempting revocation for non-existent tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.RevokeAllTenantTokensRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id  # Use real MongoDB ObjectID from setup
                ),
                target_tenant_id="000000000000000000000000"  # Non-existent tenant
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.RevokeAllTenantTokens(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")
