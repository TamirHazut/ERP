"""
Functional tests for Auth Service authentication flows.
Tests: Login, Logout, Token Refresh, Token Verification.
"""
import pytest
import grpc
import sys
import os

# Add infra functional path to sys.path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../infra/functional'))

from grpc_client import GrpcClient
from config import TestConfig
from proto.auth.v1 import auth_pb2, auth_pb2_grpc
from proto.infra.v1 import infra_pb2


@pytest.mark.auth
@pytest.mark.integration
class TestAuthenticationFlows:
    """Test authentication flows (happy path)."""

    @pytest.fixture(autouse=True)
    def setup(self, clean_database):
        """Setup test data before each test."""
        # Pre-test: Create tenant, admin user, roles, permissions
        # This would use a test data seeder or direct DB inserts
        self.tenant_id = "test-tenant-123"
        self.user_email = "testuser@example.com"
        self.user_password = "SecurePassword123!"

    def test_login_success(self):
        """Test successful user login flow."""
        # Arrange
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Act
            request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                account_id=self.user_email,
                password=self.user_password
            )

            response = stub.Login(request)

            # Assert
            assert response.tokens is not None
            assert response.tokens.token != ""
            assert response.tokens.refresh_token != ""
            assert response.expires_in.token > 0
            assert response.expires_in.refresh_token > 0

    def test_logout_success(self):
        """Test successful user logout flow."""
        # Pre-test: Login to get tokens
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Login first
            login_request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                account_id=self.user_email,
                password=self.user_password
            )
            login_response = stub.Login(login_request)
            access_token = login_response.tokens.token
            refresh_token = login_response.tokens.refresh_token

            # Act: Logout
            logout_request = auth_pb2.LogoutRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id="test-user-id"  # Would get from token verification
                ),
                access_token=access_token,
                refresh_token=refresh_token
            )

            logout_response = stub.Logout(logout_request)

            # Assert
            assert logout_response.message == "logout successful"

    def test_refresh_token_success(self):
        """Test token refresh flow."""
        # Pre-test: Login to get refresh token
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Login first
            login_request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                account_id=self.user_email,
                password=self.user_password
            )
            login_response = stub.Login(login_request)
            refresh_token = login_response.tokens.refresh_token

            # Act: Refresh token
            refresh_request = auth_pb2.RefreshTokenRequest(
                tenant_id=self.tenant_id,
                refresh_token=refresh_token
            )

            refresh_response = stub.RefreshToken(refresh_request)

            # Assert
            assert refresh_response.tokens is not None
            assert refresh_response.tokens.token != ""
            assert refresh_response.tokens.refresh_token != ""

    def test_verify_token_success(self):
        """Test token verification flow."""
        # Pre-test: Login to get access token
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = auth_pb2_grpc.AuthServiceStub(client.get_channel())

            # Login first
            login_request = auth_pb2.LoginRequest(
                tenant_id=self.tenant_id,
                account_id=self.user_email,
                password=self.user_password
            )
            login_response = stub.Login(login_request)
            access_token = login_response.tokens.token

            # Act: Verify token
            verify_request = auth_pb2.VerifyTokenRequest(
                tenant_id=self.tenant_id,
                token=access_token
            )

            verify_response = stub.VerifyToken(verify_request)

            # Assert
            assert verify_response.valid is True
            assert verify_response.claims is not None
