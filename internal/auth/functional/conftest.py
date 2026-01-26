"""
Pytest fixtures for Auth service functional tests.
"""
import pytest
import sys
import os

# Add infra functional path to sys.path for imports
infra_functional_path = os.path.join(os.path.dirname(__file__), '../../infra/functional')
sys.path.insert(0, infra_functional_path)
# Add proto path for proto imports
sys.path.insert(0, os.path.join(infra_functional_path, 'proto'))

from db.manager import DatabaseManager
from config import TestConfig
from logger import setup_logging, get_log_config_from_env, get_logger

# Module logger
logger = get_logger("fixtures")


def pytest_configure(config):
    """Configure pytest session - setup logging."""
    # Setup logging system
    setup_logging(get_log_config_from_env())
    logger.info("Pytest session started")


def pytest_runtest_logstart(nodeid, location):
    """Log test start."""
    test_logger = get_logger("tests")
    test_logger.info(f"Test started: {nodeid}")


def pytest_runtest_logfinish(nodeid, location):
    """Log test finish."""
    test_logger = get_logger("tests")
    test_logger.info(f"Test finished: {nodeid}")


@pytest.fixture(scope="session", autouse=True)
def setup_test_environment():
    """Setup test environment once per test session."""
    logger.info("Setting up test environment")
    db_manager = DatabaseManager()

    # Setup
    db_manager.setup()
    db_manager.clean_test_data()

    yield

    # Teardown
    logger.info("Tearing down test environment")
    db_manager.clean_test_data()
    db_manager.teardown()


@pytest.fixture(scope="function")
def clean_database():
    """Clean database before each test."""
    logger.info("Cleaning test databases")
    db_manager = DatabaseManager()
    db_manager.setup()
    db_manager.clean_test_data()
    yield
    # Cleanup after test if needed


@pytest.fixture(scope="session")
def test_config():
    """Provide test configuration."""
    return TestConfig
