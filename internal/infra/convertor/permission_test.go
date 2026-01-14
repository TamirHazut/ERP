package convertor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	model_auth "erp.localhost/internal/infra/model/auth"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
)

// =============================================================================
// PermissionToProto Tests
// =============================================================================

func TestPermissionToProto_ValidPermission(t *testing.T) {
	now := time.Now()
	objectID := primitive.NewObjectID()

	permission := &model_auth.Permission{
		ID:               objectID,
		TenantID:         "tenant-123",
		Resource:         model_auth.ResourceTypeUser,
		Action:           model_auth.PermissionActionCreate,
		PermissionString: "user:create",
		DisplayName:      "Create User",
		Description:      "Allows creating new users",
		CreatedBy:        "admin",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	result := PermissionToProto(permission)

	require.NotNil(t, result)
	assert.Equal(t, objectID.Hex(), result.Id)
	assert.Equal(t, "tenant-123", result.TenantId)
	assert.Equal(t, "Create User", result.Name)
	assert.Equal(t, "user:create", result.Slug)
	assert.Equal(t, "Allows creating new users", result.Description)
	assert.Equal(t, model_auth.ResourceTypeUser, result.Resource)
	assert.Equal(t, model_auth.PermissionActionCreate, result.Action)
	assert.Equal(t, model_auth.PermissionStatusActive, result.Status)
	assert.Equal(t, "admin", result.CreatedBy)
	assert.NotNil(t, result.CreatedAt)
	assert.NotNil(t, result.UpdatedAt)
}

func TestPermissionToProto_NilInput(t *testing.T) {
	result := PermissionToProto(nil)
	assert.Nil(t, result)
}

func TestPermissionToProto_InvalidPermission(t *testing.T) {
	// Permission with invalid data that will fail validation
	invalidPerm := &model_auth.Permission{
		ID:       primitive.NewObjectID(),
		TenantID: "", // Missing required field
	}

	result := PermissionToProto(invalidPerm)
	assert.Nil(t, result, "Should return nil for invalid permission")
}

// =============================================================================
// PermissionsToProto Tests
// =============================================================================

func TestPermissionsToProto_ValidSlice(t *testing.T) {
	now := time.Now()
	permissions := []model_auth.Permission{
		{
			ID:               primitive.NewObjectID(),
			TenantID:         "tenant-123",
			Resource:         model_auth.ResourceTypeUser,
			Action:           model_auth.PermissionActionCreate,
			PermissionString: "user:create",
			DisplayName:      "Create User",
			CreatedBy:        "admin",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               primitive.NewObjectID(),
			TenantID:         "tenant-123",
			Resource:         model_auth.ResourceTypeUser,
			Action:           model_auth.PermissionActionRead,
			PermissionString: "user:read",
			DisplayName:      "Read User",
			CreatedBy:        "admin",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	result := PermissionsToProto(permissions)

	require.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "Create User", result[0].Name)
	assert.Equal(t, "Read User", result[1].Name)
}

func TestPermissionsToProto_EmptySlice(t *testing.T) {
	result := PermissionsToProto([]model_auth.Permission{})

	require.NotNil(t, result)
	assert.Len(t, result, 0)
}

// =============================================================================
// CreatePermissionFromProto Tests
// =============================================================================

func TestCreatePermissionFromProto_ValidProto(t *testing.T) {
	proto := &proto_auth.CreatePermissionData{
		TenantId:    "tenant-123",
		Resource:    model_auth.ResourceTypeUser,
		Action:      model_auth.PermissionActionCreate,
		Slug:        "user:create",
		Name:        "Create User",
		Description: "Allows creating new users",
		CreatedBy:   "admin",
	}

	result, err := CreatePermissionFromProto(proto)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "tenant-123", result.TenantID)
	assert.Equal(t, model_auth.ResourceTypeUser, result.Resource)
	assert.Equal(t, model_auth.PermissionActionCreate, result.Action)
	assert.Equal(t, "user:create", result.PermissionString)
	assert.Equal(t, "Create User", result.DisplayName)
	assert.Equal(t, "Allows creating new users", result.Description)
	assert.Equal(t, "admin", result.CreatedBy)
	assert.Equal(t, "", result.Category)
	assert.Equal(t, false, result.IsDangerous)
	assert.Equal(t, false, result.RequiresApproval)
	assert.NotNil(t, result.Dependencies)
	assert.Len(t, result.Dependencies, 0)
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())
}

func TestCreatePermissionFromProto_NilProto(t *testing.T) {
	result, err := CreatePermissionFromProto(nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "proto")
}

// =============================================================================
// UpdatePermissionFromProto Tests
// =============================================================================

func TestUpdatePermissionFromProto_FullUpdate(t *testing.T) {
	existing := &model_auth.Permission{
		ID:               primitive.NewObjectID(),
		TenantID:         "tenant-123",
		Resource:         model_auth.ResourceTypeUser,
		Action:           model_auth.PermissionActionCreate,
		PermissionString: "user:create",
		DisplayName:      "Old Name",
		Description:      "Old Description",
		CreatedBy:        "admin",
		CreatedAt:        time.Now().Add(-1 * time.Hour),
		UpdatedAt:        time.Now().Add(-1 * time.Hour),
	}

	newName := "New Name"
	newSlug := "user:update"
	newDesc := "New Description"
	newResource := model_auth.ResourceTypeRole
	newAction := model_auth.PermissionActionUpdate

	proto := &proto_auth.UpdatePermissionData{
		Id:          existing.ID.Hex(),
		Name:        &newName,
		Slug:        &newSlug,
		Description: &newDesc,
		Resource:    &newResource,
		Action:      &newAction,
	}

	err := UpdatePermissionFromProto(existing, proto)

	require.NoError(t, err)
	assert.Equal(t, "New Name", existing.DisplayName)
	assert.Equal(t, "user:update", existing.PermissionString)
	assert.Equal(t, "New Description", existing.Description)
	assert.Equal(t, model_auth.ResourceTypeRole, existing.Resource)
	assert.Equal(t, model_auth.PermissionActionUpdate, existing.Action)
	// UpdatedAt should be updated to a recent time
	assert.True(t, existing.UpdatedAt.After(existing.CreatedAt))
}

func TestUpdatePermissionFromProto_PartialUpdate(t *testing.T) {
	existing := &model_auth.Permission{
		ID:               primitive.NewObjectID(),
		TenantID:         "tenant-123",
		Resource:         model_auth.ResourceTypeUser,
		Action:           model_auth.PermissionActionCreate,
		PermissionString: "user:create",
		DisplayName:      "Original Name",
		Description:      "Original Description",
		CreatedBy:        "admin",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Only update Name and Description, leave others as nil
	updatedName := "Updated Name"
	updatedDesc := "Updated Description"

	proto := &proto_auth.UpdatePermissionData{
		Id:          existing.ID.Hex(),
		Name:        &updatedName,
		Description: &updatedDesc,
		// Slug, Resource, Action are nil - should not be updated
	}

	err := UpdatePermissionFromProto(existing, proto)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", existing.DisplayName)
	assert.Equal(t, "Updated Description", existing.Description)
	// These should remain unchanged
	assert.Equal(t, "user:create", existing.PermissionString)
	assert.Equal(t, model_auth.ResourceTypeUser, existing.Resource)
	assert.Equal(t, model_auth.PermissionActionCreate, existing.Action)
}

func TestUpdatePermissionFromProto_NilExisting(t *testing.T) {
	testName := "Test"
	proto := &proto_auth.UpdatePermissionData{
		Id:   primitive.NewObjectID().Hex(),
		Name: &testName,
	}

	err := UpdatePermissionFromProto(nil, proto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "existing")
}

func TestUpdatePermissionFromProto_NilProto(t *testing.T) {
	existing := &model_auth.Permission{
		ID:       primitive.NewObjectID(),
		TenantID: "tenant-123",
	}

	err := UpdatePermissionFromProto(existing, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "proto")
}

func TestUpdatePermissionFromProto_BothNil(t *testing.T) {
	err := UpdatePermissionFromProto(nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "existing")
}
