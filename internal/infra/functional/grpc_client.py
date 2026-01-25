"""
Generic gRPC client wrapper for functional tests.
Provides reusable utilities for creating and managing gRPC connections.
"""
import grpc
from typing import Optional, Any
from .config import ServiceConfig


class GrpcClient:
    """Generic gRPC client wrapper."""

    def __init__(self, service_config: ServiceConfig):
        self.config = service_config
        self.channel: Optional[grpc.Channel] = None

    def __enter__(self):
        """Context manager entry - creates gRPC channel."""
        if self.config.use_tls:
            # TODO: Implement mTLS when certificates are ready
            raise NotImplementedError("TLS not yet implemented")
        else:
            self.channel = grpc.insecure_channel(self.config.endpoint)
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit - closes gRPC channel."""
        if self.channel:
            self.channel.close()

    def get_channel(self) -> grpc.Channel:
        """Get the gRPC channel."""
        if not self.channel:
            raise RuntimeError("Channel not initialized. Use context manager.")
        return self.channel
