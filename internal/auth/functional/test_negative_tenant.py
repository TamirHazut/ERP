"""
Functional tests for Tenant Service - Negative test cases (tenant.proto).
Tests error scenarios: duplicate names, invalid states, invalid deletions.
"""
from bson import ObjectId
from datetime import datetime
import pytest
import grpc
import sys
import os

# Add infra functional path to sys.path for imports
infra_functional_path = os.path.join(os.path.dirname(__file__), '../../infra/functional')
sys.path.insert(0, infra_functional_path)
# Add proto path for proto imports
sys.path.insert(0, os.path.join(infra_functional_path, 'proto'))
# Add auth functional path for helpers
auth_functional_path = os.path.dirname(__file__)
sys.path.insert(0, auth_functional_path)

from grpc_client import GrpcClient
from config import TestConfig
from auth.v1 import tenant_pb2, tenant_pb2_grpc
from infra.v1 import infra_pb2
from logger import get_logger
from db.mongo_client import MongoDBClient
from db.redis_client import RedisClient
from seeders.system_seeder import SystemSeeder
from helpers.db_injection import inject_tenant

# Test logger
logger = get_logger("tests.tenant.negative")


@pytest.mark.tenant
@pytest.mark.negative
@pytest.mark.integration
class TestTenantManagementErrors:
    """Test tenant management error scenarios."""

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

    def test_create_tenant_duplicate_name(self):
        """Test CreateTenant with tenant name that already exists."""
        logger.info("Step 1: Creating first tenant")

        # Pre-test: Create first tenant
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            tenant1 = tenant_pb2.Tenant(
                name="Duplicate Tenant Corp",
                slug="duplicate-tenant-1",
                status=tenant_pb2.TENANT_STATUS_ACTIVE,
                created_by=self.admin_user_id,
                contact=tenant_pb2.ContactInfo(email="test@erp.com")
            )

            create_request1 = tenant_pb2.CreateTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant=tenant1
            )
            stub.CreateTenant(create_request1)

            logger.info("Step 2: Attempting to create second tenant with same name")
            tenant2 = tenant_pb2.Tenant(
                name="Duplicate Tenant Corp",  # Same name
                slug="duplicate-tenant-2",
                status=tenant_pb2.TENANT_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request2 = tenant_pb2.CreateTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant=tenant2
            )

            logger.info("Step 3: Expecting ALREADY_EXISTS error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateTenant(create_request2)

            assert exc_info.value.code() == grpc.StatusCode.ALREADY_EXISTS
            logger.info("Step 4: Test completed - received expected ALREADY_EXISTS error")

    def test_create_tenant_empty_name(self):
        """Test CreateTenant with missing tenant name."""
        logger.info("Step 1: Attempting to create tenant with empty name")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            tenant = tenant_pb2.Tenant(
                name="",  # Empty name
                slug="empty-tenant",
                status=tenant_pb2.TENANT_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            create_request = tenant_pb2.CreateTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant=tenant
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateTenant(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_create_tenant_invalid_status(self):
        """Test CreateTenant with invalid status value."""
        logger.info("Step 1: Attempting to create tenant with invalid status")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            tenant = tenant_pb2.Tenant(
                name="Invalid Status Tenant",
                slug="invalid-status-tenant",
                status=999,  # Invalid status
                created_by=self.admin_user_id
            )

            create_request = tenant_pb2.CreateTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant=tenant
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.CreateTenant(create_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_get_tenant_nonexistent_by_id(self):
        """Test GetTenant with invalid tenant_id."""
        logger.info("Step 1: Attempting to get tenant with non-existent ID")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            request = tenant_pb2.GetTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant_id="000000000000000000000000"  # Non-existent tenant
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetTenant(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_tenant_nonexistent_by_name(self):
        """Test GetTenant with non-existent name."""
        logger.info("Step 1: Attempting to get tenant with non-existent name")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            request = tenant_pb2.GetTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                name="Non Existent Tenant Name"
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetTenant(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_get_tenant_missing_identifier(self):
        """Test GetTenant with empty ID and name."""
        logger.info("Step 1: Attempting to get tenant without ID or name")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            # Create request with neither tenant_id nor name
            request = tenant_pb2.GetTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id)
                # Missing both tenant_id and name
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.GetTenant(request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")

    def test_list_tenants_invalid_pagination(self):
        """Test ListTenants with invalid pagination (page_size = 0)."""
        logger.info("Step 1: Attempting to list tenants with invalid pagination")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            request = tenant_pb2.ListTenantsRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id)
                # Note: pagination parameter with page_size = 0 would go here if proto supports it
            )

            logger.info("Step 2: Note - Skipping test as pagination validation may not be implemented yet")
            pytest.skip("Pagination validation may not be implemented in ListTenants endpoint yet")

    def test_update_tenant_nonexistent(self):
        """Test UpdateTenant with invalid tenant_id."""
        logger.info("Step 1: Attempting to update non-existent tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            tenant = tenant_pb2.Tenant(
                id="000000000000000000000000",  # Non-existent tenant
                name="Updated Tenant",
                slug="updated-tenant",
                status=tenant_pb2.TENANT_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            request = tenant_pb2.UpdateTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant=tenant
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdateTenant(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_update_tenant_duplicate_name(self):
        """Test UpdateTenant changing to existing name."""
        logger.info("Step 1: Injecting two tenants directly into MongoDB")

        # Pre-test: Inject two tenants directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            # Inject tenant1
            inject_tenant(
                mongo,
                name="Tenant One",
                slug="tenant-one",
                status=1,  # TENANT_STATUS_ACTIVE
                contact={"email": "test@erp.com"},
                created_by=self.admin_user_id
            )

            # Inject tenant2
            tenant2_id = inject_tenant(
                mongo,
                name="Tenant Two",
                slug="tenant-two",
                status=1,  # TENANT_STATUS_ACTIVE
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting to update tenant2 name to tenant1's name")
        # Act - Test UpdateTenant gRPC endpoint with duplicate name
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            tenant2 = tenant_pb2.Tenant(
                id=tenant2_id,
                name="Tenant One",  # Duplicate name
                slug="tenant-two",
                status=tenant_pb2.TENANT_STATUS_ACTIVE,
                created_by=self.admin_user_id
            )

            update_request = tenant_pb2.UpdateTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant=tenant2
            )

            logger.info("Step 3: Expecting ALREADY_EXISTS error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdateTenant(update_request)

            assert exc_info.value.code() == grpc.StatusCode.ALREADY_EXISTS
            logger.info("Step 4: Test completed - received expected ALREADY_EXISTS error")

    def test_update_tenant_invalid_status_transition(self):
        """Test UpdateTenant with invalid state transition."""
        logger.info("Step 1: Injecting tenant in TRIAL status directly into MongoDB")

        # Pre-test: Inject tenant with TRIAL status directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            tenant_id = inject_tenant(
                mongo,
                name="Trial Tenant",
                slug="trial-tenant",
                status=2,  # TENANT_STATUS_TRIAL
                contact={"email": "test@erp.com"},
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Attempting invalid status transition TRIAL → SUSPENDED")
        # Act - Test UpdateTenant gRPC endpoint with invalid status transition
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            tenant = tenant_pb2.Tenant(
                id=tenant_id,
                name="Trial Tenant",
                slug="trial-tenant",
                status=tenant_pb2.TENANT_STATUS_SUSPENDED,
                created_by=self.admin_user_id,
                contact=tenant_pb2.ContactInfo(email="test@erp.com")
            )

            update_request = tenant_pb2.UpdateTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant=tenant
            )

            logger.info("Step 3: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.UpdateTenant(update_request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 4: Test completed - received expected INVALID_ARGUMENT error")

    def test_delete_tenant_nonexistent(self):
        """Test DeleteTenant with non-existent tenant_id."""
        logger.info("Step 1: Attempting to delete non-existent tenant")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            request = tenant_pb2.DeleteTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant_id="000000000000000000000000"  # Non-existent tenant
            )

            logger.info("Step 2: Expecting NOT_FOUND error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.DeleteTenant(request)

            assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND
            logger.info("Step 3: Test completed - received expected NOT_FOUND error")

    def test_delete_tenant_missing_id(self):
        """Test DeleteTenant with empty tenant_id."""
        logger.info("Step 1: Attempting to delete tenant with empty tenant_id")

        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            request = tenant_pb2.DeleteTenantRequest(
                identifier=infra_pb2.UserIdentifier(tenant_id=self.tenant_id, user_id=self.admin_user_id),
                tenant_id=""  # Empty tenant_id
            )

            logger.info("Step 2: Expecting INVALID_ARGUMENT error")
            with pytest.raises(grpc.RpcError) as exc_info:
                stub.DeleteTenant(request)

            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
            logger.info("Step 3: Test completed - received expected INVALID_ARGUMENT error")
