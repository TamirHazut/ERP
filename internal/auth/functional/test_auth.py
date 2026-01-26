"""
Functional tests for Auth Service (auth.proto).
Tests: Login, Logout, Token Refresh, Token Verification, Token Revocation.
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
from auth.v1 import auth_pb2, auth_pb2_grpc
from infra.v1 import infra_pb2
from logger import get_logger

# Test logger
logger = get_logger("tests.auth")


@pytest.mark.auth
@pytest.mark.integration
class TestAuthenticationFlows:
    """Test authentication flows (happy path)."""

    @pytest.fixture(autouse=True)
    def setup(self, clean_database):
        """Setup test data before each test."""
        # Import seeder
        from db.mongo_client import MongoDBClient
        from seeders.system_seeder import SystemSeeder

        database=os.getenv("AUTH_DB_NAME","auth_db_test")  # Separate test database

        # Seed system data (tenant, permission, role, admin user)
        with MongoDBClient(database) as mongo:
            seeder = SystemSeeder(mongo)
            system_data = seeder.seed_all()

            # Store IDs for use in tests (these are real MongoDB ObjectIDs)
            self.tenant_id = system_data["tenant_id"]
            self.user_id = system_data["user_id"]
            self.role_id = system_data["role_id"]
            self.permission_id = system_data["permission_id"]

        # Use default credentials from config
        self.user_email = TestConfig.DEFAULT_ADMIN_EMAIL
        self.user_password = TestConfig.DEFAULT_ADMIN_PASSWORD

    def test_login_success(self):
        """Test successful user login flow."""
        logger.info(f"Step 1: Creating login request for user: {self.user_email}")

        # Arrange
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            logger.info("Step 2: Sending Login RPC request")
            # Act
            request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            )

            response = stub.Login(request)

            logger.info("Step 3: Validating response - checking tokens received")
            # Assert
            assert response.tokens is not None
            assert response.tokens.token != ""

            logger.info("Step 4: Validating response - checking refresh token received")
            assert response.tokens.refresh_token != ""

            logger.info("Step 5: Validating response - checking token expiration times")
            assert response.expires_in.token > 0
            assert response.expires_in.refresh_token > 0

            logger.info("Step 6: Login test completed successfully")

    def test_logout_success(self):
        """Test successful user logout flow."""
        logger.info("Step 1: Logging in to obtain tokens for logout test")

        # Pre-test: Login to get tokens
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Login first
            login_request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            )
            login_response = stub.Login(login_request)
            token = login_response.tokens.token
            refresh_token = login_response.tokens.refresh_token

            logger.info("Step 2: Creating logout request with obtained tokens")
            # Act: Logout
            logout_request = auth_pb2.LogoutRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id  # Use real MongoDB ObjectID from setup
                ),
                tokens=auth_pb2.Tokens(
                    token=token,
                    refresh_token=refresh_token
                )
            )

            logger.info("Step 3: Sending Logout RPC request")
            logout_response = stub.Logout(logout_request)

            logger.info("Step 4: Validating logout response message")
            # Assert
            assert logout_response.message == "logout successful"

            logger.info("Step 5: Logout test completed successfully")

    def test_refresh_token_success(self):
        """Test token refresh flow."""
        logger.info("Step 1: Logging in to obtain refresh token")

        # Pre-test: Login to get refresh token
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Login first
            login_request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            )
            login_response = stub.Login(login_request)
            refresh_token = login_response.tokens.refresh_token

            logger.info("Step 2: Creating refresh token request")
            # Act: Refresh token
            refresh_request = auth_pb2.RefreshTokenRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id  # Use real MongoDB ObjectID from setup
                ),
                refresh_token=refresh_token
            )

            logger.info("Step 3: Sending RefreshToken RPC request")
            refresh_response = stub.RefreshToken(refresh_request)

            logger.info("Step 4: Validating new tokens received")
            # Assert
            assert refresh_response.tokens is not None
            assert refresh_response.tokens.token != ""
            assert refresh_response.tokens.refresh_token != ""

            logger.info("Step 5: Token refresh test completed successfully")

    def test_verify_token_success(self):
        """Test token verification flow."""
        logger.info("Step 1: Logging in to obtain access token")

        # Pre-test: Login to get access token
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Login first
            login_request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            )
            login_response = stub.Login(login_request)
            access_token = login_response.tokens.token

            logger.info("Step 2: Creating token verification request")
            # Act: Verify token
            verify_request = auth_pb2.VerifyTokenRequest(
                token=access_token
            )

            logger.info("Step 3: Sending VerifyToken RPC request")
            verify_response = stub.VerifyToken(verify_request)

            logger.info("Step 4: Validating token is valid")
            # Assert
            assert verify_response.valid is True

            logger.info("Step 5: Token verification test completed successfully")


@pytest.mark.auth
@pytest.mark.integration
class TestTokenRevocation:
    """Test token revocation flows (happy path)."""

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

    def test_revoke_token_success(self):
        """Test revoking a user's access token."""
        logger.info("Step 1: Pre-test - logging in to obtain tokens")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Pre-test: Login to get tokens
            login_request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                email=self.user_email,
                password=self.user_password
            )
            login_response = stub.Login(login_request)
            access_token = login_response.tokens.token
            refresh_token = login_response.tokens.refresh_token

            logger.info("Step 2: Act - calling RevokeToken RPC")
            # Act: Revoke the tokens
            revoke_request = auth_pb2.RevokeTokenRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                revoked_by=self.user_id,
                tokens=auth_pb2.Tokens(
                    token=access_token,
                    refresh_token=refresh_token
                )
            )
            revoke_response = stub.RevokeToken(revoke_request)

            logger.info("Step 3: Assert - validating token was revoked")
            # Assert
            assert revoke_response.revoked is True

            logger.info("Step 4: Verify - checking revoked token is invalid")
            # Verify: Check that the revoked token is now invalid
            verify_request = auth_pb2.VerifyTokenRequest(token=access_token)
            verify_response = stub.VerifyToken(verify_request)
            assert verify_response.valid is False

            logger.info("Step 5: Token revocation test completed successfully")

    def test_revoke_all_tenant_tokens_success(self):
        """Test revoking all tokens for a tenant."""
        logger.info("Step 1: Pre-test - logging in 3 times to create multiple token sets")

        tokens = []
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Pre-test: Login 3 times to create 3 sets of tokens
            for i in range(3):
                login_request = auth_pb2.LoginRequest(
                    tenant_id=self.tenant_id,
                    email=self.user_email,
                    password=self.user_password
                )
                login_response = stub.Login(login_request)
                tokens.append({
                    'access': login_response.tokens.token,
                    'refresh': login_response.tokens.refresh_token
                })
                logger.info(f"Created token set {i+1}/3")

            logger.info("Step 2: Act - calling RevokeAllTenantTokens RPC")
            # Act: Revoke all tokens for the tenant
            revoke_all_request = auth_pb2.RevokeAllTenantTokensRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.user_id
                ),
                target_tenant_id=self.tenant_id
            )
            revoke_all_response = stub.RevokeAllTenantTokens(revoke_all_request)

            logger.info("Step 3: Assert - validating revocation response")
            # Assert
            assert revoke_all_response.revoked is True
            assert revoke_all_response.access_tokens_revoked >= 3
            assert revoke_all_response.refresh_tokens_revoked >= 3
            logger.info(f"Revoked {revoke_all_response.access_tokens_revoked} access tokens and {revoke_all_response.refresh_tokens_revoked} refresh tokens")

            logger.info("Step 4: Verify - checking all access tokens are invalid")
            # Verify: All access tokens should now be invalid
            for i, token_set in enumerate(tokens):
                verify_request = auth_pb2.VerifyTokenRequest(token=token_set['access'])
                verify_response = stub.VerifyToken(verify_request)
                assert verify_response.valid is False
                logger.info(f"Token {i+1}/3 is invalid (as expected)")

            logger.info("Step 5: RevokeAllTenantTokens test completed successfully")
