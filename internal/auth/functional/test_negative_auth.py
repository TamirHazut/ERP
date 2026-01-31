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

from grpc_client import GrpcClient
from bson import ObjectId
from config import TestConfig
from auth.v1 import auth_pb2, auth_pb2_grpc, user_pb2, user_pb2_grpc
from infra.v1 import infra_pb2
from logger import get_logger
from db.mongo_client import MongoDBClient
from db.redis_client import RedisClient
from seeders.system_seeder import SystemSeeder

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

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Login(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

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

            logger.info("Step 3: Expecting PERMISSION_DENIED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Login(request)

            assert exc_info.value.code() == grpc.StatusCode.PERMISSION_DENIED
            logger.info("Step 4: Test completed - received expected PERMISSION_DENIED error")

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

    def test_logout_invalid_token(self):
        """Test logout with malformed token."""
        logger.info("Step 1: Attempting logout with malformed token")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.LogoutRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id  # Use real MongoDB ObjectID from setup
                ),
                tokens=auth_pb2.Tokens(
                    token="invalid access token",
                    refresh_token="invalid refresh token"
                )
            )

            logger.info("Step 2: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Logout(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 3: Test completed - received expected UNAUTHENTICATED error")

    def test_logout_revoked_token(self):
        """Test logout with already revoked token."""
        logger.info("Step 1: Logging in to get token")

        # Pre-test: Login and revoke token
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Login
            login_request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            )
            login_response = stub.Login(login_request)
            token = login_response.tokens.token
            refresh_roken = login_response.tokens.refresh_token

            # Logout to revoke token
            logger.info("Step 2: Logging out to revoke token")
            logout_request = auth_pb2.LogoutRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id  # Use real MongoDB ObjectID from setup
                ),
                tokens=auth_pb2.Tokens(
                    token=token,
                    refresh_token=refresh_roken
                ))
            stub.Logout(logout_request)

            # Try to logout again with same token
            logger.info("Step 3: Attempting logout with already revoked token")
            logger.info("Step 4: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.Logout(logout_request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 5: Test completed - received expected UNAUTHENTICATED error")

    def test_logout_expired_token(self):
        """Test logout with expired access token."""
        logger.info("Step 1: Note - This test requires token to naturally expire or mock time")
        logger.info("Skipping implementation - requires time manipulation or very short token TTL")
        pytest.skip("Requires time manipulation or very short token TTL configuration")

    def test_refresh_expired_refresh_token(self):
        """Test RefreshToken with expired refresh token."""
        logger.info("Step 1: Note - This test requires refresh token to naturally expire")
        logger.info("Skipping implementation - requires time manipulation or very short refresh token TTL")
        pytest.skip("Requires time manipulation or very short refresh token TTL configuration")

    def test_refresh_invalid_refresh_token(self):
        """Test RefreshToken with malformed refresh token."""
        logger.info("Step 1: Attempting token refresh with malformed refresh token")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.RefreshTokenRequest(
                refresh_token="invalid.refresh.token"
            )

            logger.info("Step 2: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.RefreshToken(request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 3: Test completed - received expected UNAUTHENTICATED error")

    def test_refresh_revoked_refresh_token(self):
        """Test RefreshToken with revoked refresh token."""
        logger.info("Step 1: Logging in to get refresh token")

        # Pre-test: Login and revoke all tokens
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Login
            login_request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            )
            login_response = stub.Login(login_request)
            refresh_token = login_response.tokens.refresh_token

            # Revoke all tokens for user
            logger.info("Step 2: Revoking all user tokens")
            revoke_request = auth_pb2.RevokeTokenRequest(
                tenant_id=self.tenant_id,
                user_id=self.admin_user_id
            )
            stub.RevokeToken(revoke_request)

            # Try to refresh with revoked token
            logger.info("Step 3: Attempting refresh with revoked refresh token")
            refresh_request = auth_pb2.RefreshTokenRequest(
                refresh_token=refresh_token
            )

            logger.info("Step 4: Expecting UNAUTHENTICATED error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.RefreshToken(refresh_request)

            assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED
            logger.info("Step 5: Test completed - received expected UNAUTHENTICATED error")

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
        """Test VerifyToken with expired access token."""
        logger.info("Step 1: Note - This test requires token to naturally expire or mock time")
        logger.info("Skipping implementation - requires time manipulation or very short token TTL")
        pytest.skip("Requires time manipulation or very short token TTL configuration")

    def test_revoke_nonexistent_user(self):
        """Test RevokeToken with invalid user_id."""
        logger.info("Step 1: Attempting token revocation for non-existent user")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.RevokeTokenRequest(
                tenant_id=self.tenant_id,
                user_id="000000000000000000000000"  # Non-existent user ID
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.RevokeToken(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_revoke_all_tenant_invalid_tenant(self):
        """Test RevokeAllTenantTokens with non-existent tenant."""
        logger.info("Step 1: Attempting revocation for non-existent tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            request = auth_pb2.RevokeAllTenantTokensRequest(
                tenant_id=self.tenant_id,
                user_id=self.admin_user_id,
                target_tenant_id="000000000000000000000000"  # Non-existent tenant
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.RevokeAllTenantTokens(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")
