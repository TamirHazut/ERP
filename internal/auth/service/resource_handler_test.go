package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	auth_models "erp.localhost/internal/infra/model/auth"
	auth_proto "erp.localhost/internal/infra/proto/auth/v1"
)

// =============================================================================
// Permission Handler Tests
// =============================================================================

func TestPermissionHandler_ExtractAndConvertCreate_Success(t *testing.T) {
	handler := &permissionHandler{}

	req := &auth_proto.CreateResourceRequest{
		Resource: &auth_proto.CreateResourceRequest_Permission{
			Permission: &auth_proto.CreatePermissionData{
				TenantId:    "tenant-123",
				Resource:    auth_models.ResourceTypeUser,
				Action:      auth_models.PermissionActionCreate,
				Slug:        "user:create",
				Name:        "Create User",
				Description: "Allows creating users",
				CreatedBy:   "admin",
			},
		},
	}

	result, err := handler.ExtractAndConvertCreate(req)

	require.NoError(t, err)
	require.NotNil(t, result)

	perm, ok := result.(*auth_models.Permission)
	require.True(t, ok, "Result should be *auth_models.Permission")
	assert.Equal(t, "tenant-123", perm.TenantID)
	assert.Equal(t, auth_models.ResourceTypeUser, perm.Resource)
	assert.Equal(t, auth_models.PermissionActionCreate, perm.Action)
	assert.Equal(t, "user:create", perm.PermissionString)
	assert.Equal(t, "Create User", perm.DisplayName)
}

func TestPermissionHandler_ExtractAndConvertCreate_InvalidType(t *testing.T) {
	handler := &permissionHandler{}

	// Wrong type - Role instead of Permission
	req := &auth_proto.CreateResourceRequest{
		Resource: &auth_proto.CreateResourceRequest_Role{
			Role: &auth_proto.CreateRoleData{
				TenantId: "tenant-123",
				Name:     "Admin",
			},
		},
	}

	result, err := handler.ExtractAndConvertCreate(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "resource")
}

func TestPermissionHandler_ExtractAndConvertCreate_ConversionError(t *testing.T) {
	handler := &permissionHandler{}

	// Nil permission data will cause conversion error
	req := &auth_proto.CreateResourceRequest{
		Resource: &auth_proto.CreateResourceRequest_Permission{
			Permission: nil,
		},
	}

	result, err := handler.ExtractAndConvertCreate(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "permission")
}

func TestPermissionHandler_ExtractUpdateData_Success(t *testing.T) {
	handler := &permissionHandler{}

	updateName := "Updated Name"
	updateProto := &auth_proto.UpdatePermissionData{
		Id:   primitive.NewObjectID().Hex(),
		Name: &updateName,
	}

	req := &auth_proto.UpdateResourceRequest{
		Resource: &auth_proto.UpdateResourceRequest_Permission{
			Permission: updateProto,
		},
	}

	result, err := handler.ExtractUpdateData(req)

	require.NoError(t, err)
	require.NotNil(t, result)

	data, ok := result.(*auth_proto.UpdatePermissionData)
	require.True(t, ok)
	assert.Equal(t, updateProto, data)
}

func TestPermissionHandler_ExtractUpdateData_InvalidType(t *testing.T) {
	handler := &permissionHandler{}

	// Wrong type - Role instead of Permission
	req := &auth_proto.UpdateResourceRequest{
		Resource: &auth_proto.UpdateResourceRequest_Role{
			Role: &auth_proto.UpdateRoleData{
				Id: primitive.NewObjectID().Hex(),
			},
		},
	}

	result, err := handler.ExtractUpdateData(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "resource")
}

func TestPermissionHandler_GetResourceIDFromUpdate(t *testing.T) {
	handler := &permissionHandler{}

	objectID := primitive.NewObjectID()
	updateData := &auth_proto.UpdatePermissionData{
		Id: objectID.Hex(),
	}

	result := handler.GetResourceIDFromUpdate(updateData)

	assert.Equal(t, objectID.Hex(), result)
}

func TestPermissionHandler_ApplyUpdate_Success(t *testing.T) {
	handler := &permissionHandler{}

	existing := &auth_models.Permission{
		ID:               primitive.NewObjectID(),
		TenantID:         "tenant-123",
		Resource:         auth_models.ResourceTypeUser,
		Action:           auth_models.PermissionActionCreate,
		PermissionString: "user:create",
		DisplayName:      "Old Name",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	newName := "New Name"
	updateData := &auth_proto.UpdatePermissionData{
		Id:   existing.ID.Hex(),
		Name: &newName,
	}

	err := handler.ApplyUpdate(existing, updateData)

	require.NoError(t, err)
	assert.Equal(t, "New Name", existing.DisplayName)
}

func TestPermissionHandler_ApplyUpdate_InvalidExistingType(t *testing.T) {
	handler := &permissionHandler{}

	// Wrong type - Role instead of Permission
	existing := &auth_models.Role{
		ID:       primitive.NewObjectID(),
		TenantID: "tenant-123",
		Name:     "Admin",
	}

	updateData := &auth_proto.UpdatePermissionData{
		Id: primitive.NewObjectID().Hex(),
	}

	err := handler.ApplyUpdate(existing, updateData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "existing resource")
}

func TestPermissionHandler_ApplyUpdate_InvalidUpdateDataType(t *testing.T) {
	handler := &permissionHandler{}

	existing := &auth_models.Permission{
		ID:       primitive.NewObjectID(),
		TenantID: "tenant-123",
	}

	// Wrong type - Role update data instead of Permission
	updateData := &auth_proto.UpdateRoleData{
		Id: primitive.NewObjectID().Hex(),
	}

	err := handler.ApplyUpdate(existing, updateData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update data")
}

func TestPermissionHandler_GetResourceType(t *testing.T) {
	handler := &permissionHandler{}

	result := handler.GetResourceType()

	assert.Equal(t, auth_models.ResourceTypePermission, result)
}

func TestPermissionHandler_GetResourceName(t *testing.T) {
	handler := &permissionHandler{}

	result := handler.GetResourceName()

	assert.Equal(t, "permission", result)
}

// =============================================================================
// Role Handler Tests
// =============================================================================

func TestRoleHandler_ExtractAndConvertCreate_Success(t *testing.T) {
	handler := &roleHandler{}

	req := &auth_proto.CreateResourceRequest{
		Resource: &auth_proto.CreateResourceRequest_Role{
			Role: &auth_proto.CreateRoleData{
				TenantId:    "tenant-123",
				Name:        "Admin",
				Slug:        "admin",
				Description: "Administrator role",
				Type:        "system",
				CreatedBy:   "system",
			},
		},
	}

	result, err := handler.ExtractAndConvertCreate(req)

	require.NoError(t, err)
	require.NotNil(t, result)

	role, ok := result.(*auth_models.Role)
	require.True(t, ok, "Result should be *auth_models.Role")
	assert.Equal(t, "tenant-123", role.TenantID)
	assert.Equal(t, "Admin", role.Name)
	assert.Equal(t, "admin", role.Slug)
	assert.Equal(t, "Administrator role", role.Description)
}

func TestRoleHandler_ExtractAndConvertCreate_InvalidType(t *testing.T) {
	handler := &roleHandler{}

	// Wrong type - Permission instead of Role
	req := &auth_proto.CreateResourceRequest{
		Resource: &auth_proto.CreateResourceRequest_Permission{
			Permission: &auth_proto.CreatePermissionData{
				TenantId: "tenant-123",
				Name:     "Create User",
			},
		},
	}

	result, err := handler.ExtractAndConvertCreate(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "resource")
}

func TestRoleHandler_ExtractAndConvertCreate_ConversionError(t *testing.T) {
	handler := &roleHandler{}

	// Nil role data will cause conversion error
	req := &auth_proto.CreateResourceRequest{
		Resource: &auth_proto.CreateResourceRequest_Role{
			Role: nil,
		},
	}

	result, err := handler.ExtractAndConvertCreate(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "role")
}

func TestRoleHandler_ExtractUpdateData_Success(t *testing.T) {
	handler := &roleHandler{}

	updateName := "Updated Role"
	updateProto := &auth_proto.UpdateRoleData{
		Id:   primitive.NewObjectID().Hex(),
		Name: &updateName,
	}

	req := &auth_proto.UpdateResourceRequest{
		Resource: &auth_proto.UpdateResourceRequest_Role{
			Role: updateProto,
		},
	}

	result, err := handler.ExtractUpdateData(req)

	require.NoError(t, err)
	require.NotNil(t, result)

	data, ok := result.(*auth_proto.UpdateRoleData)
	require.True(t, ok)
	assert.Equal(t, updateProto, data)
}

func TestRoleHandler_ExtractUpdateData_InvalidType(t *testing.T) {
	handler := &roleHandler{}

	// Wrong type - Permission instead of Role
	req := &auth_proto.UpdateResourceRequest{
		Resource: &auth_proto.UpdateResourceRequest_Permission{
			Permission: &auth_proto.UpdatePermissionData{
				Id: primitive.NewObjectID().Hex(),
			},
		},
	}

	result, err := handler.ExtractUpdateData(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "resource")
}

func TestRoleHandler_GetResourceIDFromUpdate(t *testing.T) {
	handler := &roleHandler{}

	objectID := primitive.NewObjectID()
	updateData := &auth_proto.UpdateRoleData{
		Id: objectID.Hex(),
	}

	result := handler.GetResourceIDFromUpdate(updateData)

	assert.Equal(t, objectID.Hex(), result)
}

func TestRoleHandler_ApplyUpdate_Success(t *testing.T) {
	handler := &roleHandler{}

	existing := &auth_models.Role{
		ID:          primitive.NewObjectID(),
		TenantID:    "tenant-123",
		Name:        "Old Name",
		Slug:        "old-slug",
		Description: "Old Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	newName := "New Name"
	updateData := &auth_proto.UpdateRoleData{
		Id:   existing.ID.Hex(),
		Name: &newName,
	}

	err := handler.ApplyUpdate(existing, updateData)

	require.NoError(t, err)
	assert.Equal(t, "New Name", existing.Name)
}

func TestRoleHandler_ApplyUpdate_InvalidExistingType(t *testing.T) {
	handler := &roleHandler{}

	// Wrong type - Permission instead of Role
	existing := &auth_models.Permission{
		ID:       primitive.NewObjectID(),
		TenantID: "tenant-123",
	}

	updateData := &auth_proto.UpdateRoleData{
		Id: primitive.NewObjectID().Hex(),
	}

	err := handler.ApplyUpdate(existing, updateData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "existing resource")
}

func TestRoleHandler_ApplyUpdate_InvalidUpdateDataType(t *testing.T) {
	handler := &roleHandler{}

	existing := &auth_models.Role{
		ID:       primitive.NewObjectID(),
		TenantID: "tenant-123",
	}

	// Wrong type - Permission update data instead of Role
	updateData := &auth_proto.UpdatePermissionData{
		Id: primitive.NewObjectID().Hex(),
	}

	err := handler.ApplyUpdate(existing, updateData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update data")
}

func TestRoleHandler_GetResourceType(t *testing.T) {
	handler := &roleHandler{}

	result := handler.GetResourceType()

	assert.Equal(t, auth_models.ResourceTypeRole, result)
}

func TestRoleHandler_GetResourceName(t *testing.T) {
	handler := &roleHandler{}

	result := handler.GetResourceName()

	assert.Equal(t, "role", result)
}

// =============================================================================
// getResourceHandler Tests
// =============================================================================

func TestGetResourceHandler_PermissionType(t *testing.T) {
	handler, err := getResourceHandler(auth_proto.ResourceType_RESOURCE_TYPE_PERMISSION)

	require.NoError(t, err)
	require.NotNil(t, handler)

	_, ok := handler.(*permissionHandler)
	assert.True(t, ok, "Should return permissionHandler")
	assert.Equal(t, "permission", handler.GetResourceName())
}

func TestGetResourceHandler_RoleType(t *testing.T) {
	handler, err := getResourceHandler(auth_proto.ResourceType_RESOURCE_TYPE_ROLE)

	require.NoError(t, err)
	require.NotNil(t, handler)

	_, ok := handler.(*roleHandler)
	assert.True(t, ok, "Should return roleHandler")
	assert.Equal(t, "role", handler.GetResourceName())
}

func TestGetResourceHandler_InvalidType(t *testing.T) {
	handler, err := getResourceHandler(auth_proto.ResourceType_RESOURCE_TYPE_UNSPECIFIED)

	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.Contains(t, err.Error(), "resourceType")
}

func TestGetResourceHandler_UnknownType(t *testing.T) {
	// Use a type value that doesn't exist in the registry
	handler, err := getResourceHandler(auth_proto.ResourceType(999))

	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.Contains(t, err.Error(), "resourceType")
}
