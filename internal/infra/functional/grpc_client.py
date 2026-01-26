"""
Generic gRPC client wrapper for functional tests.
Provides reusable utilities for creating and managing gRPC connections.
"""
import grpc
from typing import Optional, Any
import sys
import os

# Add current directory to path for imports
sys.path.insert(0, os.path.dirname(__file__))
from config import ServiceConfig
from logger import get_logger

# Module logger
logger = get_logger("grpc")


class GrpcClient:
    """Generic gRPC client wrapper."""

    def __init__(self, service_config: ServiceConfig):
        self.config = service_config
        self.channel: Optional[grpc.Channel] = None

    def __enter__(self):
        """Context manager entry - creates gRPC channel."""
        try:
            if self.config.use_tls:
                # TODO: Implement mTLS when certificates are ready
                raise NotImplementedError("TLS not yet implemented")
            else:
                self.channel = grpc.insecure_channel(self.config.endpoint)
            logger.info(f"Created gRPC channel: {self.config.endpoint}")
            return self
        except Exception as e:
            logger.error(f"Failed to create gRPC channel to {self.config.endpoint}: {e}")
            raise

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit - closes gRPC channel."""
        if self.channel:
            self.channel.close()
            logger.info(f"Closed gRPC channel: {self.config.endpoint}")

    def get_channel(self) -> grpc.Channel:
        """Get the gRPC channel."""
        if not self.channel:
            raise RuntimeError("Channel not initialized. Use context manager.")
        return self.channel
