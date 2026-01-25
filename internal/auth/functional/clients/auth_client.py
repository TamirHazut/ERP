"""
Auth Service gRPC client wrapper for functional tests.
Provides high-level methods for authentication operations.
"""
import sys
import os

# Add proto path to system path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../infra/functional'))

from proto.auth.v1 import auth_pb2, auth_pb2_grpc
from proto.infra.v1 import infra_pb2
import grpc
from typing import Optional, Tuple


class AuthClient:
    """High-level client for Auth Service."""

    def __init__(self, channel: grpc.Channel):
        self.stub = auth_pb2_grpc.AuthServiceStub(channel)

    def login(self, tenant_id: str, account_id: str, password: str) -> Tuple[str, str, int, int]:
        """
        Login user and get tokens.

        Returns: (access_token, refresh_token, access_expiry, refresh_expiry)
        """
        request = auth_pb2.LoginRequest(
            tenant_id=tenant_id,
            account_id=account_id,
            password=password
        )
        response = self.stub.Login(request)

        return (
            response.tokens.token,
            response.tokens.refresh_token,
            response.expires_in.token,
            response.expires_in.refresh_token
        )

    def logout(self, tenant_id: str, user_id: str, access_token: str, refresh_token: str) -> str:
        """
        Logout user and revoke tokens.

        Returns: Logout message
        """
        request = auth_pb2.LogoutRequest(
            identifier=infra_pb2.UserIdentifier(
                tenant_id=tenant_id,
                user_id=user_id
            ),
            access_token=access_token,
            refresh_token=refresh_token
        )
        response = self.stub.Logout(request)
        return response.message

    def refresh_token(self, tenant_id: str, refresh_token: str) -> Tuple[str, str]:
        """
        Refresh access token.

        Returns: (new_access_token, new_refresh_token)
        """
        request = auth_pb2.RefreshTokenRequest(
            tenant_id=tenant_id,
            refresh_token=refresh_token
        )
        response = self.stub.RefreshToken(request)

        return (
            response.tokens.token,
            response.tokens.refresh_token
        )

    def verify_token(self, tenant_id: str, token: str) -> bool:
        """
        Verify if a token is valid.

        Returns: True if valid, False otherwise
        """
        request = auth_pb2.VerifyTokenRequest(
            tenant_id=tenant_id,
            token=token
        )
        response = self.stub.VerifyToken(request)
        return response.valid
