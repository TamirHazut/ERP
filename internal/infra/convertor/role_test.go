package convertor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	auth_models "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/proto/auth/v1"
)

// =============================================================================
// RoleToProto Tests
// =============================================================================

func TestRoleToProto_ValidRole(t *testing.T) {
	now := time.Now()
	objectID := primitive.NewObjectID()

	role := &auth_models.Role{
		ID:           objectID,
		TenantID:     "tenant-123",
		Name:         "Admin",
		Slug:         "admin",
		Description:  "Administrator role",
		Type:         "system",
		IsSystemRole: true,
		Permissions:  []string{"user:create", "user:read"},
		Priority:     10,
		Status:       "active",
		CreatedBy:    "system",
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata: auth_models.RoleMetadata{
			Color:         "#FF0000",
			Icon:          "admin-icon",
			MaxAssignable: 5,
		},
	}

	result := RoleToProto(role)

	require.NotNil(t, result)
	assert.Equal(t, objectID.Hex(), result.Id)
	assert.Equal(t, "tenant-123", result.TenantId)
	assert.Equal(t, "Admin", result.Name)
	assert.Equal(t, "admin", result.Slug)
	assert.Equal(t, "Administrator role", result.Description)
	assert.Equal(t, "system", result.Type)
	assert.Equal(t, true, result.IsSystemRole)
	assert.Equal(t, []string{"user:create", "user:read"}, result.Permissions)
	assert.Equal(t, int32(10), result.Priority)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "system", result.CreatedBy)
	assert.NotNil(t, result.CreatedAt)
	assert.NotNil(t, result.UpdatedAt)
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, "#FF0000", result.Metadata.Color)
	assert.Equal(t, "admin-icon", result.Metadata.Icon)
	assert.Equal(t, int32(5), result.Metadata.MaxAssignable)
}

func TestRoleToProto_NilInput(t *testing.T) {
	result := RoleToProto(nil)
	assert.Nil(t, result)
}

func TestRoleToProto_InvalidRole(t *testing.T) {
	invalidRole := &auth_models.Role{
		ID:       primitive.NewObjectID(),
		TenantID: "", // Missing required field
	}

	result := RoleToProto(invalidRole)
	assert.Nil(t, result, "Should return nil for invalid role")
}

// =============================================================================
// RoleMetadataToProto Tests
// =============================================================================

func TestRoleMetadataToProto_ValidMetadata(t *testing.T) {
	metadata := &auth_models.RoleMetadata{
		Color:         "#00FF00",
		Icon:          "user-icon",
		MaxAssignable: 10,
	}

	result := RoleMetadataToProto(metadata)

	require.NotNil(t, result)
	assert.Equal(t, "#00FF00", result.Color)
	assert.Equal(t, "user-icon", result.Icon)
	assert.Equal(t, int32(10), result.MaxAssignable)
}

func TestRoleMetadataToProto_NilMetadata(t *testing.T) {
	result := RoleMetadataToProto(nil)
	assert.Nil(t, result)
}

// =============================================================================
// RolesToProto Tests
// =============================================================================

func TestRolesToProto_ValidSlice(t *testing.T) {
	now := time.Now()
	roles := []auth_models.Role{
		{
			ID:          primitive.NewObjectID(),
			TenantID:    "tenant-123",
			Name:        "Admin",
			Slug:        "admin",
			Description: "Admin role",
			Status:      "active",
			Permissions: []string{"user:read", "user:write"},
			CreatedBy:   "system",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			TenantID:    "tenant-123",
			Name:        "User",
			Slug:        "user",
			Description: "User role",
			Status:      "active",
			Permissions: []string{"user:read"},
			CreatedBy:   "system",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	result := RolesToProto(roles)

	require.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "Admin", result[0].Name)
	assert.Equal(t, "User", result[1].Name)
}

func TestRolesToProto_EmptySlice(t *testing.T) {
	result := RolesToProto([]auth_models.Role{})

	require.NotNil(t, result)
	assert.Len(t, result, 0)
}

// =============================================================================
// CreateRoleFromProto Tests
// =============================================================================

func TestCreateRoleFromProto_ValidProto(t *testing.T) {
	proto := &authv1.CreateRoleData{
		TenantId:     "tenant-123",
		Name:         "Editor",
		Slug:         "editor",
		Description:  "Editor role",
		Type:         "custom",
		IsSystemRole: false,
		Permissions:  []string{"user:read", "user:update"},
		Priority:     5,
		Status:       "active",
		CreatedBy:    "admin",
	}

	result, err := CreateRoleFromProto(proto)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "tenant-123", result.TenantID)
	assert.Equal(t, "Editor", result.Name)
	assert.Equal(t, "editor", result.Slug)
	assert.Equal(t, "Editor role", result.Description)
	assert.Equal(t, "custom", result.Type)
	assert.Equal(t, false, result.IsSystemRole)
	assert.Equal(t, []string{"user:read", "user:update"}, result.Permissions)
	assert.Equal(t, 5, result.Priority)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "admin", result.CreatedBy)
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())
}

func TestCreateRoleFromProto_WithMetadata(t *testing.T) {
	proto := &authv1.CreateRoleData{
		TenantId: "tenant-123",
		Name:     "Manager",
		Slug:     "manager",
		Metadata: &authv1.RoleMetadata{
			Color:         "#0000FF",
			Icon:          "manager-icon",
			MaxAssignable: 3,
		},
		CreatedBy: "admin",
	}

	result, err := CreateRoleFromProto(proto)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "#0000FF", result.Metadata.Color)
	assert.Equal(t, "manager-icon", result.Metadata.Icon)
	assert.Equal(t, 3, result.Metadata.MaxAssignable)
}

func TestCreateRoleFromProto_WithoutMetadata(t *testing.T) {
	proto := &authv1.CreateRoleData{
		TenantId:  "tenant-123",
		Name:      "Basic",
		Slug:      "basic",
		CreatedBy: "admin",
	}

	result, err := CreateRoleFromProto(proto)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "", result.Metadata.Color)
	assert.Equal(t, "", result.Metadata.Icon)
	assert.Equal(t, 0, result.Metadata.MaxAssignable)
}

func TestCreateRoleFromProto_NilProto(t *testing.T) {
	result, err := CreateRoleFromProto(nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "proto")
}

// =============================================================================
// UpdateRoleFromProto Tests
// =============================================================================

func TestUpdateRoleFromProto_FullUpdate(t *testing.T) {
	existing := &auth_models.Role{
		ID:           primitive.NewObjectID(),
		TenantID:     "tenant-123",
		Name:         "Old Name",
		Slug:         "old-slug",
		Description:  "Old Description",
		Type:         "old-type",
		IsSystemRole: false,
		Permissions:  []string{"old:perm"},
		Priority:     1,
		Status:       "inactive",
		CreatedBy:    "admin",
		CreatedAt:    time.Now().Add(-1 * time.Hour),
		UpdatedAt:    time.Now().Add(-1 * time.Hour),
	}

	newName := "New Name"
	newSlug := "new-slug"
	newDesc := "New Description"
	newType := "new-type"
	newSystemRole := true
	newPriority := int32(10)
	newStatus := "active"

	proto := &authv1.UpdateRoleData{
		Id:           existing.ID.Hex(),
		Name:         &newName,
		Slug:         &newSlug,
		Description:  &newDesc,
		Type:         &newType,
		IsSystemRole: &newSystemRole,
		Permissions:  []string{"new:perm1", "new:perm2"},
		Priority:     &newPriority,
		Status:       &newStatus,
	}

	err := UpdateRoleFromProto(existing, proto)

	require.NoError(t, err)
	assert.Equal(t, "New Name", existing.Name)
	assert.Equal(t, "new-slug", existing.Slug)
	assert.Equal(t, "New Description", existing.Description)
	assert.Equal(t, "new-type", existing.Type)
	assert.Equal(t, true, existing.IsSystemRole)
	assert.Equal(t, []string{"new:perm1", "new:perm2"}, existing.Permissions)
	assert.Equal(t, 10, existing.Priority)
	assert.Equal(t, "active", existing.Status)
	assert.True(t, existing.UpdatedAt.After(existing.CreatedAt))
}

func TestUpdateRoleFromProto_PartialUpdate(t *testing.T) {
	existing := &auth_models.Role{
		ID:          primitive.NewObjectID(),
		TenantID:    "tenant-123",
		Name:        "Original Name",
		Slug:        "original-slug",
		Description: "Original Description",
		Type:        "original-type",
		Priority:    5,
		CreatedBy:   "admin",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	newName := "Updated Name"
	newDesc := "Updated Description"

	proto := &authv1.UpdateRoleData{
		Id:          existing.ID.Hex(),
		Name:        &newName,
		Description: &newDesc,
		// Other fields are nil - should not be updated
	}

	err := UpdateRoleFromProto(existing, proto)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", existing.Name)
	assert.Equal(t, "Updated Description", existing.Description)
	// These should remain unchanged
	assert.Equal(t, "original-slug", existing.Slug)
	assert.Equal(t, "original-type", existing.Type)
	assert.Equal(t, 5, existing.Priority)
}

func TestUpdateRoleFromProto_NilExisting(t *testing.T) {
	testName := "Test"
	proto := &authv1.UpdateRoleData{
		Id:   primitive.NewObjectID().Hex(),
		Name: &testName,
	}

	err := UpdateRoleFromProto(nil, proto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "existing")
}

func TestUpdateRoleFromProto_NilProto(t *testing.T) {
	existing := &auth_models.Role{
		ID:       primitive.NewObjectID(),
		TenantID: "tenant-123",
	}

	err := UpdateRoleFromProto(existing, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "proto")
}

func TestUpdateRoleFromProto_BothNil(t *testing.T) {
	err := UpdateRoleFromProto(nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "existing")
}

// =============================================================================
// RoleObjectIDFromString Tests
// =============================================================================

func TestRoleObjectIDFromString_ValidHex(t *testing.T) {
	validID := primitive.NewObjectID()
	hexString := validID.Hex()

	result, err := RoleObjectIDFromString(hexString)

	require.NoError(t, err)
	assert.Equal(t, validID, result)
}

func TestRoleObjectIDFromString_EmptyString(t *testing.T) {
	result, err := RoleObjectIDFromString("")

	assert.Error(t, err)
	assert.Equal(t, primitive.NilObjectID, result)
	assert.Contains(t, err.Error(), "id")
}

func TestRoleObjectIDFromString_InvalidHex(t *testing.T) {
	result, err := RoleObjectIDFromString("invalid-hex-string")

	assert.Error(t, err)
	assert.Equal(t, primitive.NilObjectID, result)
}
