"""
Database injection helpers for functional tests.
Inject test data directly into MongoDB to avoid using gRPC stubs for test setup.
"""

from datetime import datetime
from bson import ObjectId


def inject_user(mongo_client, tenant_id, email, username, status=1, roles=None, **kwargs):
    """
    Inject user document directly into MongoDB.

    Args:
        mongo_client: MongoDB client instance
        tenant_id: Tenant ID for the user
        email: User email
        username: Username
        status: User status (1=ACTIVE, 2=INACTIVE, 3=SUSPENDED)
        roles: List of role assignments (optional)
        **kwargs: Additional fields (password_hash, profile, preferences, created_by, etc.)

    Returns:
        str: Inserted user ID
    """
    user_doc = {
        "_id": ObjectId(),
        "tenant_id": tenant_id,
        "email": email,
        "username": username,
        "password_hash": kwargs.get("password_hash", "hashed_password_default"),
        "profile": kwargs.get("profile", {
            "first_name": username.capitalize(),
            "last_name": "User",
            "display_name": f"{username.capitalize()} User"
        }),
        "roles": roles or [],
        "additional_permissions": [],
        "revoked_permissions": [],
        "status": status,
        "email_verified": kwargs.get("email_verified", False),
        "phone_verified": kwargs.get("phone_verified", False),
        "mfa_enabled": kwargs.get("mfa_enabled", False),
        "preferences": kwargs.get("preferences", {
            "language": "en",
            "timezone": "UTC",
            "theme": "light",
            "notifications": {"email": True, "push": False, "sms": False}
        }),
        "created_at": datetime.now(),
        "updated_at": datetime.now(),
        "created_by": kwargs.get("created_by", "system")
    }

    users_collection = mongo_client.get_collection("users")
    result = users_collection.insert_one(user_doc)
    return str(result.inserted_id)


def inject_tenant(mongo_client, name, slug, status=1, **kwargs):
    """
    Inject tenant document directly into MongoDB.

    Args:
        mongo_client: MongoDB client instance
        name: Tenant name
        slug: Tenant slug
        status: Tenant status (1=ACTIVE, 2=TRIAL, 3=SUSPENDED, 4=CANCELLED)
        **kwargs: Additional fields (subscription, settings, contact, created_by, etc.)

    Returns:
        str: Inserted tenant ID
    """
    tenant_doc = {
        "_id": ObjectId(),
        "name": name,
        "slug": slug,
        "status": status,
        "subscription": kwargs.get("subscription", {
            "plan": "basic",
            "limits": {
                "max_users": 10,
                "max_products": 100,
                "max_orders_per_month": 50,
                "storage_gb": 5
            }
        }),
        "settings": kwargs.get("settings", {
            "timezone": "UTC",
            "currency": "USD",
            "date_format": "YYYY-MM-DD",
            "language": "en"
        }),
        "contact": kwargs.get("contact", {
            "email": f"admin@{slug}.com",
            "phone": "+1-555-0100"
        }),
        "created_at": datetime.now(),
        "updated_at": datetime.now(),
        "created_by": kwargs.get("created_by", "system")
    }

    tenants_collection = mongo_client.get_collection("tenants")
    result = tenants_collection.insert_one(tenant_doc)
    return str(result.inserted_id)


def inject_role(mongo_client, tenant_id, name, permissions, **kwargs):
    """
    Inject role document directly into MongoDB.

    Args:
        mongo_client: MongoDB client instance
        tenant_id: Tenant ID for the role
        name: Role name
        permissions: List of permission IDs
        **kwargs: Additional fields (description, type, status, created_by, etc.)

    Returns:
        str: Inserted role ID
    """
    role_doc = {
        "_id": ObjectId(),
        "tenant_id": tenant_id,
        "name": name,
        "description": kwargs.get("description", f"{name} role"),
        "type": kwargs.get("type", 2),  # ROLE_TYPE_CUSTOM = 2, ROLE_TYPE_SYSTEM = 1
        "permissions": permissions,
        "status": kwargs.get("status", 1),  # ROLE_STATUS_ACTIVE = 1
        "protected": kwargs.get("protected", False),
        "created_at": datetime.now(),
        "updated_at": datetime.now(),
        "created_by": kwargs.get("created_by", "system")
    }

    roles_collection = mongo_client.get_collection("roles")
    result = roles_collection.insert_one(role_doc)
    return str(result.inserted_id)


def inject_permission(mongo_client, tenant_id, permission_string, resource, action, **kwargs):
    """
    Inject permission document directly into MongoDB.

    Args:
        mongo_client: MongoDB client instance
        tenant_id: Tenant ID for the permission
        permission_string: Permission string (e.g., "user:read")
        resource: Resource name (e.g., "users")
        action: Action name (e.g., "read")
        **kwargs: Additional fields (display_name, description, status, is_dangerous, created_by, etc.)

    Returns:
        str: Inserted permission ID
    """
    permission_doc = {
        "_id": ObjectId(),
        "tenant_id": tenant_id,
        "display_name": kwargs.get("display_name", f"{resource.capitalize()} {action.capitalize()}"),
        "permission_string": permission_string,
        "description": kwargs.get("description", f"Permission for {permission_string}"),
        "resource": resource,
        "action": action,
        "status": kwargs.get("status", 1),  # PERMISSION_STATUS_ACTIVE = 1
        "is_dangerous": kwargs.get("is_dangerous", False),
        "protected": kwargs.get("protected", False),
        "created_at": datetime.now(),
        "updated_at": datetime.now(),
        "created_by": kwargs.get("created_by", "system")
    }

    permissions_collection = mongo_client.get_collection("permissions")
    result = permissions_collection.insert_one(permission_doc)
    return str(result.inserted_id)


def inject_tenant_with_defaults(mongo_client, name, slug, admin_user_id, **kwargs):
    """
    Inject tenant with automatic seeding of defaults (permission, role, user).

    This mimics what CreateTenant gRPC does automatically:
    1. Creates wildcard permission (*:*)
    2. Creates TenantAdmin role
    3. Creates admin user for tenant

    Use this when tests expect the full tenant setup with defaults.

    Args:
        mongo_client: MongoDB client instance
        name: Tenant name
        slug: Tenant slug
        admin_user_id: User ID to use as creator and assign admin role
        **kwargs: Additional fields passed to inject_tenant

    Returns:
        dict: Dictionary with tenant_id, permission_id, role_id, user_id
    """
    # Inject tenant
    tenant_id = inject_tenant(mongo_client, name, slug, **kwargs)

    # Inject wildcard permission
    perm_id = inject_permission(
        mongo_client, tenant_id, "*:*", "*", "*",
        display_name="Full Access",
        description="Grants full access to all resources",
        is_dangerous=True,
        created_by=admin_user_id
    )

    # Inject TenantAdmin role
    role_id = inject_role(
        mongo_client, tenant_id, "TenantAdmin", [perm_id],
        description="Tenant administrator with full access",
        type=1,  # ROLE_TYPE_SYSTEM
        created_by=admin_user_id
    )

    # Inject admin user
    user_id = inject_user(
        mongo_client, tenant_id, f"admin@{slug}.com", "admin",
        password_hash="hashed_password_admin",
        profile={
            "first_name": "Admin",
            "last_name": "User",
            "display_name": "Admin User"
        },
        roles=[{
            "tenant_id": tenant_id,
            "role_id": role_id,
            "assigned_at": datetime.now(),
            "assigned_by": admin_user_id
        }],
        created_by=admin_user_id
    )

    return {
        "tenant_id": tenant_id,
        "permission_id": perm_id,
        "role_id": role_id,
        "user_id": user_id
    }
