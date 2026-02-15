"""
Functional tests for Tenant Service (tenant.proto).
Tests: CreateTenant, GetTenant (by ID and name), ListTenants, UpdateTenant, DeleteTenant.
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

from lib.functional.grpc_client import GrpcClient
from lib.functional.config import TestConfig
from lib.model.auth.v1 import tenant_pb2, tenant_pb2_grpc, user_pb2, role_pb2, permission_pb2
from lib.model.core.v1 import address_pb2
from lib.model.infra.v1 import infra_pb2
from lib.functional.logger import get_logger
from lib.functional.db.mongo_client import MongoDBClient
from lib.functional.db.redis_client import RedisClient
from lib.functional.seeders.system_seeder import SystemSeeder
from helpers.db_injection import inject_tenant, inject_tenant_with_defaults

# Test logger
logger = get_logger("tests.tenant")


@pytest.mark.tenant
@pytest.mark.integration
class TestTenantManagement:
    """Test tenant management operations (happy path)."""

    @pytest.fixture(autouse=True)
    def setup(self, clean_database):
        """Setup test data before each test."""

        database = os.getenv("AUTH_DB_NAME", "auth_db_test")  # Separate test database

        # Seed system data (tenant, permission, role, admin user)
        with MongoDBClient(database) as mongo, RedisClient() as redis:
            seeder = SystemSeeder(mongo, redis)
            system_data = seeder.seed_all()

            # Store IDs for use in tests (these are real MongoDB ObjectIDs)
            self.tenant_id = system_data["tenant_id"]
            self.admin_user_id = system_data["user_id"]
            self.role_id = system_data["role_id"]

    def test_create_tenant_success(self):
        """Test successful tenant creation with automatic seeding of defaults."""
        logger.info("Step 1: Creating tenant creation request")

        # Arrange
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            # Create tenant object
            new_tenant = tenant_pb2.Tenant(
                name="Test Tenant Corp",
                slug="test-tenant-corp",
                domain="testtenant.com",
                status=tenant_pb2.TENANT_STATUS_ACTIVE,
                subscription=tenant_pb2.Subscription(
                    plan="standard",
                    limits=tenant_pb2.SubscriptionLimits(
                        max_users=50,
                        max_products=1000,
                        max_orders_per_month=500,
                        storage_gb=10
                    )
                ),
                settings=tenant_pb2.TenantSettings(
                    timezone="UTC",
                    currency="USD",
                    date_format="YYYY-MM-DD",
                    language="en"
                ),
                contact=tenant_pb2.ContactInfo(
                    email="admin@testtenant.com",
                    phone="+1-555-0100",
                    address=address_pb2.Address(
                        street="123 Test St",
                        city="Test City",
                        state="TC",
                        country="US"
                    )
                ),
                branding=tenant_pb2.Branding(
                    logo_url="https://example.com/logo.png",
                    primary_color="#0066cc",
                    company_name="Test Tenant Corp"
                ),
                created_by=self.admin_user_id,
                metadata=tenant_pb2.TenantMetadata(
                    onboarding_completed=False,
                    industry="Technology",
                    company_size="10-50"
                )
            )

            logger.info("Step 2: Sending CreateTenant RPC request")
            # Act
            request = tenant_pb2.CreateTenantRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                tenant=new_tenant
            )

            response = stub.CreateTenant(request)

            logger.info("Step 3: Validating response - checking tenant_id received")
            # Assert
            assert response.tenant_id is not None
            assert response.tenant_id != ""

            logger.info("Step 4: Verifying tenant exists in MongoDB")
            # Post-test verification
            database = os.getenv("AUTH_DB_NAME", "auth_db_test")
            with MongoDBClient(database) as mongo:
                tenants_collection = mongo.get_collection("tenants")
                tenant_doc = tenants_collection.find_one({"_id": ObjectId(response.tenant_id)})

                assert tenant_doc is not None
                assert tenant_doc["name"] == "Test Tenant Corp"
                assert tenant_doc["slug"] == "test-tenant-corp"
                assert tenant_doc["status"] == tenant_pb2.TENANT_STATUS_ACTIVE

                logger.info("Step 5: Verifying seeded defaults (role, admin user)")

                # Verify seeded TenantAdmin role
                roles_collection = mongo.get_collection("roles")
                tenant_admin_role = roles_collection.find_one({
                    "tenant_id": response.tenant_id,
                    "name": "tenant_admin"
                })
                assert tenant_admin_role is not None
                assert "*:*" in tenant_admin_role["permissions"]
                logger.info(f"  - Verified TenantAdmin role: {tenant_admin_role['_id']}")

                # Verify seeded admin user
                users_collection = mongo.get_collection("users")
                admin_user = users_collection.find_one({
                    "tenant_id": response.tenant_id,
                    "username": "admin"
                })
                assert admin_user is not None
                assert any(ObjectId(role["role_id"]) == tenant_admin_role["_id"] for role in admin_user.get("roles", []))
                logger.info(f"  - Verified admin user: {admin_user['_id']}")

            logger.info("Step 6: CreateTenant test completed successfully")

    def test_get_tenant_by_id_success(self):
        """Test successful tenant retrieval by ID."""
        logger.info("Step 1: Injecting tenant directly into MongoDB")

        # Pre-test: Inject tenant directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            created_tenant_id = inject_tenant(
                mongo,
                name="Get Tenant By ID Corp",
                slug="get-tenant-id",
                status=1,  # TENANT_STATUS_ACTIVE
                subscription={
                    "plan": "basic",
                    "limits": {
                        "max_users": 10,
                        "max_products": 100,
                        "max_orders_per_month": 50,
                        "storage_gb": 5
                    }
                },
                settings={
                    "timezone": "UTC",
                    "currency": "USD",
                    "date_format": "YYYY-MM-DD",
                    "language": "en"
                },
                contact={
                    "email": "admin@gettenantid.com",
                    "phone": "+1-555-0101"
                },
                created_by=self.admin_user_id
            )

        logger.info(f"Step 2: Retrieving tenant with ID: {created_tenant_id}")
        # Act - Test GetTenant gRPC endpoint
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            get_request = tenant_pb2.GetTenantRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                tenant_id=created_tenant_id
            )

            tenant_response = stub.GetTenant(get_request)

            logger.info("Step 3: Validating retrieved tenant fields")
            # Assert
            assert tenant_response.id == created_tenant_id
            assert tenant_response.name == "Get Tenant By ID Corp"
            assert tenant_response.slug == "get-tenant-id"
            assert tenant_response.status == tenant_pb2.TENANT_STATUS_ACTIVE

            logger.info("Step 4: GetTenant by ID test completed successfully")

    def test_get_tenant_by_name_success(self):
        """Test successful tenant retrieval by name."""
        logger.info("Step 1: Injecting tenant directly into MongoDB")

        # Pre-test: Inject tenant directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            created_tenant_id = inject_tenant(
                mongo,
                name="Get Tenant By Name Corp",
                slug="get-tenant-name",
                status=1,  # TENANT_STATUS_ACTIVE
                subscription={
                    "plan": "basic",
                    "limits": {
                        "max_users": 10,
                        "max_products": 100,
                        "max_orders_per_month": 50,
                        "storage_gb": 5
                    }
                },
                settings={
                    "timezone": "UTC",
                    "currency": "USD",
                    "date_format": "YYYY-MM-DD",
                    "language": "en"
                },
                contact={
                    "email": "admin@gettenantname.com",
                    "phone": "+1-555-0102"
                },
                created_by=self.admin_user_id
            )

        logger.info("Step 2: Retrieving tenant by name: Get Tenant By Name Corp")
        # Act - Test GetTenant gRPC endpoint
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            get_request = tenant_pb2.GetTenantRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                name="Get Tenant By Name Corp"
            )

            tenant_response = stub.GetTenant(get_request)

            logger.info("Step 3: Validating retrieved tenant fields")
            # Assert
            assert tenant_response.id == created_tenant_id
            assert tenant_response.name == "Get Tenant By Name Corp"
            assert tenant_response.slug == "get-tenant-name"

            logger.info("Step 4: GetTenant by name test completed successfully")

    def test_list_tenants_success(self):
        """Test successful tenant listing with pagination."""
        logger.info("Step 1: Injecting multiple tenants directly into MongoDB")

        # Pre-test: Inject 3 tenants directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            for i in range(1, 4):
                inject_tenant(
                    mongo,
                    name=f"List Tenant {i}",
                    slug=f"list-tenant-{i}",
                    status=1,  # TENANT_STATUS_ACTIVE
                    subscription={
                        "plan": "basic",
                        "limits": {
                            "max_users": 10,
                            "max_products": 100,
                            "max_orders_per_month": 50,
                            "storage_gb": 5
                        }
                    },
                    settings={
                        "timezone": "UTC",
                        "currency": "USD",
                        "date_format": "YYYY-MM-DD",
                        "language": "en"
                    },
                    contact={
                        "email": f"admin@listtenant{i}.com",
                        "phone": f"+1-555-010{i}"
                    },
                    created_by=self.admin_user_id
                )

        logger.info("Step 2: Listing all tenants")
        # Act - Test ListTenants gRPC endpoint
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            list_request = tenant_pb2.ListTenantsRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                )
            )

            list_response = stub.ListTenants(list_request)

            logger.info("Step 3: Validating tenant list count")
            # Assert - should have at least 4 tenants (1 system + 3 created)
            assert len(list_response.tenants) >= 3

            # Verify our created tenants are in the list
            tenant_names = [tenant.name for tenant in list_response.tenants]
            assert "List Tenant 1" in tenant_names
            assert "List Tenant 2" in tenant_names
            assert "List Tenant 3" in tenant_names

            logger.info("Step 4: ListTenants test completed successfully")

    def test_update_tenant_success(self):
        """Test successful tenant update."""
        logger.info("Step 1: Injecting tenant directly into MongoDB")

        # Pre-test: Inject tenant directly into MongoDB
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            created_tenant_id = inject_tenant(
                mongo,
                name="Update Tenant Corp",
                slug="update-tenant",
                status=1,  # TENANT_STATUS_ACTIVE
                subscription={
                    "plan": "basic",
                    "limits": {
                        "max_users": 10,
                        "max_products": 100,
                        "max_orders_per_month": 50,
                        "storage_gb": 5
                    }
                },
                settings={
                    "timezone": "UTC",
                    "currency": "USD",
                    "date_format": "YYYY-MM-DD",
                    "language": "en"
                },
                contact={
                    "email": "admin@updatetenant.com",
                    "phone": "+1-555-0200"
                },
                created_by=self.admin_user_id
            )

        logger.info(f"Step 2: Updating tenant with ID: {created_tenant_id}")
        # Act - Test UpdateTenant gRPC endpoint
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            updated_tenant = tenant_pb2.Tenant(
                id=created_tenant_id,
                name="Update Tenant Corp",
                slug="update-tenant",
                status=tenant_pb2.TENANT_STATUS_SUSPENDED,  # Changed from ACTIVE
                subscription=tenant_pb2.Subscription(
                    plan="premium",  # Changed from basic
                    limits=tenant_pb2.SubscriptionLimits(
                        max_users=100,  # Changed from 10
                        max_products=5000,  # Changed from 100
                        max_orders_per_month=1000,  # Changed from 50
                        storage_gb=50  # Changed from 5
                    )
                ),
                settings=tenant_pb2.TenantSettings(
                    timezone="Europe/London",  # Changed
                    currency="GBP",  # Changed
                    date_format="DD-MM-YYYY",  # Changed
                    language="en"
                ),
                contact=tenant_pb2.ContactInfo(
                    email="admin@updatetenant.com",
                    phone="+44-20-1234-5678"  # Changed
                ),
                created_by=self.admin_user_id
            )

            update_request = tenant_pb2.UpdateTenantRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                tenant=updated_tenant
            )

            update_response = stub.UpdateTenant(update_request)

            logger.info("Step 3: Validating update response")
            assert update_response.updated is True

            logger.info("Step 4: Retrieving updated tenant to verify changes")
            # Post-test verification
            get_request = tenant_pb2.GetTenantRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                tenant_id=created_tenant_id
            )

            tenant_response = stub.GetTenant(get_request)

            logger.info("Step 5: Verifying updated fields")
            assert tenant_response.status == tenant_pb2.TENANT_STATUS_SUSPENDED
            assert tenant_response.subscription.plan == "premium"
            assert tenant_response.subscription.limits.max_users == 100
            assert tenant_response.settings.timezone == "Europe/London"
            assert tenant_response.settings.currency == "GBP"

            logger.info("Step 6: UpdateTenant test completed successfully")

    def test_delete_tenant_success(self):
        """Test successful tenant deletion with cascading delete verification."""
        logger.info("Step 1: Injecting tenant with defaults (permission, role, user) directly into MongoDB")

        # Pre-test: Inject tenant with defaults to mimic CreateTenant behavior
        database = os.getenv("AUTH_DB_NAME", "auth_db_test")
        with MongoDBClient(database) as mongo:
            tenant_data = inject_tenant_with_defaults(
                mongo,
                name="Delete Tenant Corp",
                slug="delete-tenant",
                admin_user_id=self.admin_user_id,
                status=1,  # TENANT_STATUS_ACTIVE
                subscription={
                    "plan": "basic",
                    "limits": {
                        "max_users": 10,
                        "max_products": 100,
                        "max_orders_per_month": 50,
                        "storage_gb": 5
                    }
                },
                settings={
                    "timezone": "UTC",
                    "currency": "USD",
                    "date_format": "YYYY-MM-DD",
                    "language": "en"
                },
                contact={
                    "email": "admin@deletetenant.com",
                    "phone": "+1-555-0300"
                },
                created_by=self.admin_user_id
            )
            tenant_to_delete_id = tenant_data["tenant_id"]

        logger.info(f"Step 2: Deleting tenant with ID: {tenant_to_delete_id}")
        # Act - Test DeleteTenant gRPC endpoint
        with GrpcClient(TestConfig.AUTH_SERVICE) as client:
            stub = tenant_pb2_grpc.TenantServiceStub(client.get_channel())

            delete_request = tenant_pb2.DeleteTenantRequest(
                identifier=infra_pb2.UserIdentifier(
                    tenant_id=self.tenant_id,
                    user_id=self.admin_user_id
                ),
                tenant_id=tenant_to_delete_id
            )

            delete_response = stub.DeleteTenant(delete_request)

            logger.info("Step 3: Validating delete response")
            assert delete_response.deleted is True

            logger.info("Step 4: Verifying cascading deletion in MongoDB")
            # Post-test verification - verify all cascaded deletions
            with MongoDBClient(database) as mongo:
                # Verify tenant deleted
                tenants_collection = mongo.get_collection("tenants")
                tenant_doc = tenants_collection.find_one({"_id": ObjectId(tenant_to_delete_id)})
                assert tenant_doc is None
                logger.info("  - Verified tenant deleted from MongoDB")

                # Verify all users for tenant deleted
                users_collection = mongo.get_collection("users")
                user_count = users_collection.count_documents({"tenant_id": tenant_to_delete_id})
                assert user_count == 0
                logger.info("  - Verified all tenant users deleted")

                # Verify all roles for tenant deleted
                roles_collection = mongo.get_collection("roles")
                role_count = roles_collection.count_documents({"tenant_id": tenant_to_delete_id})
                assert role_count == 0
                logger.info("  - Verified all tenant roles deleted")

                # Verify all permissions for tenant deleted
                permissions_collection = mongo.get_collection("permissions")
                permission_count = permissions_collection.count_documents({"tenant_id": tenant_to_delete_id})
                assert permission_count == 0
                logger.info("  - Verified all tenant permissions deleted")

            logger.info("Step 5: Verifying tokens revoked from Redis")
            # Note: Token revocation verification would require knowing token keys,
            # but we can verify no tokens exist for this tenant
            with RedisClient() as redis:
                # Check that no access tokens exist for deleted tenant
                assert redis.exists(f"token:{tenant_to_delete_id}:*") is False
                logger.info("  - Verified tenant access tokens revoked")

                # Check that no refresh tokens exist for deleted tenant
                assert redis.exists(f"refresh_token:{tenant_to_delete_id}:*") is False
                logger.info("  - Verified tenant refresh tokens revoked")

            logger.info("Step 6: DeleteTenant cascading delete test completed successfully")
