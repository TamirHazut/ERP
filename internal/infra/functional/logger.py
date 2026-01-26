"""
Module-level logger for functional tests.

This module provides a centralized logging system for all functional tests.
It integrates with pytest's logging system and supports configuration via
environment variables.

Usage:
    from logger import get_logger

    logger = get_logger("db.mongo")
    logger.info("Connected to MongoDB")
    logger.debug("Operation: insert_one, collection: users")
"""

import logging
import os
from dataclasses import dataclass
from typing import Optional


@dataclass
class LogConfig:
    """Configuration for logging system."""

    level: str = "INFO"
    format: str = "[%(asctime)s] [%(name)s] [%(levelname)s] %(message)s"
    date_format: str = "%Y-%m-%d %H:%M:%S"
    file_path: Optional[str] = None


def get_log_config_from_env() -> LogConfig:
    """
    Get logging configuration from environment variables.

    Environment variables:
        LOG_LEVEL: Logging level (DEBUG, INFO, WARNING, ERROR) - default: INFO
        LOG_FILE: Optional file output path - default: None (console only)
        LOG_FORMAT: Custom format string - default: standard format

    Returns:
        LogConfig: Logging configuration
    """
    level = os.environ.get("LOG_LEVEL", "INFO").upper()
    file_path = os.environ.get("LOG_FILE")
    log_format = os.environ.get(
        "LOG_FORMAT",
        "[%(asctime)s] [%(name)s] [%(levelname)s] %(message)s"
    )

    return LogConfig(
        level=level,
        format=log_format,
        file_path=file_path
    )


def setup_logging(config: Optional[LogConfig] = None) -> None:
    """
    Set up the logging system for functional tests.

    This function should be called once at the start of the test session,
    typically from pytest_configure() hook in conftest.py.

    Args:
        config: Logging configuration. If None, uses environment variables.
    """
    if config is None:
        config = get_log_config_from_env()

    # Get the root functional logger
    root_logger = logging.getLogger("functional")

    # Set level
    root_logger.setLevel(getattr(logging, config.level))

    # Clear any existing handlers
    root_logger.handlers.clear()

    # Create formatter
    formatter = logging.Formatter(
        fmt=config.format,
        datefmt=config.date_format
    )

    # Add console handler (always)
    console_handler = logging.StreamHandler()
    console_handler.setLevel(getattr(logging, config.level))
    console_handler.setFormatter(formatter)
    root_logger.addHandler(console_handler)

    # Add file handler (if configured)
    if config.file_path:
        file_handler = logging.FileHandler(config.file_path, mode='a')
        file_handler.setLevel(getattr(logging, config.level))
        file_handler.setFormatter(formatter)
        root_logger.addHandler(file_handler)

    # Prevent propagation to root logger (avoid duplicate logs in pytest)
    root_logger.propagate = False


def get_logger(name: str) -> logging.Logger:
    """
    Get a logger with hierarchical naming.

    The logger name will be prefixed with "functional." to create
    a hierarchical structure. For example:
        get_logger("db.mongo") -> "functional.db.mongo"
        get_logger("grpc") -> "functional.grpc"

    Args:
        name: Logger name (without "functional." prefix)

    Returns:
        logging.Logger: Configured logger
    """
    return logging.getLogger(f"functional.{name}")
