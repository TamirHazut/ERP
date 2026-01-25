"""
Pytest fixtures for Auth service functional tests.
"""
import pytest
import sys
import os

# Add infra functional path to sys.path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../infra/functional'))

from db.manager import DatabaseManager
from config import TestConfig


@pytest.fixture(scope="session", autouse=True)
def setup_test_environment():
    """Setup test environment once per test session."""
    db_manager = DatabaseManager()

    # Setup
    db_manager.setup()
    db_manager.clean_test_data()

    yield

    # Teardown
    db_manager.clean_test_data()
    db_manager.teardown()


@pytest.fixture(scope="function")
def clean_database():
    """Clean database before each test."""
    db_manager = DatabaseManager()
    db_manager.setup()
    db_manager.clean_test_data()
    yield
    # Cleanup after test if needed


@pytest.fixture(scope="session")
def test_config():
    """Provide test configuration."""
    return TestConfig
