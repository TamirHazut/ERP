package convertor

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	erp_errors "erp.localhost/internal/infra/error"
	auth_models "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/proto/auth/v1"
)

// =============================================================================
// Domain Model → Proto (for responses)
// =============================================================================

// RoleToProto converts internal Role model to gRPC RoleData
func RoleToProto(role *auth_models.Role) *authv1.RoleData {
	if role == nil || role.Validate(false) != nil {
		return nil
	}

	return &authv1.RoleData{
		Id:           role.ID.Hex(),
		TenantId:     role.TenantID,
		Name:         role.Name,
		Slug:         role.Slug,
		Description:  role.Description,
		Type:         role.Type,
		IsSystemRole: role.IsSystemRole,
		Permissions:  role.Permissions,
		Priority:     int32(role.Priority),
		Status:       role.Status,
		Metadata:     RoleMetadataToProto(&role.Metadata),
		CreatedAt:    timestamppb.New(role.CreatedAt),
		UpdatedAt:    timestamppb.New(role.UpdatedAt),
		CreatedBy:    role.CreatedBy,
	}
}

// RoleMetadataToProto converts metadata
func RoleMetadataToProto(metadata *auth_models.RoleMetadata) *authv1.RoleMetadata {
	if metadata == nil {
		return nil
	}

	return &authv1.RoleMetadata{
		Color:         metadata.Color,
		Icon:          metadata.Icon,
		MaxAssignable: int32(metadata.MaxAssignable),
	}
}

// RolesToProto converts slice of roles to proto
func RolesToProto(roles []auth_models.Role) []*authv1.RoleData {
	result := make([]*authv1.RoleData, len(roles))
	for i, role := range roles {
		result[i] = RoleToProto(&role)
	}
	return result
}

// =============================================================================
// Proto → Domain Model (for create operations)
// =============================================================================

// CreateRoleFromProto converts gRPC CreateRoleData to internal Role model
func CreateRoleFromProto(proto *authv1.CreateRoleData) (*auth_models.Role, error) {
	if proto == nil {
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue, "proto")
	}

	role := &auth_models.Role{
		TenantID:     proto.TenantId,
		Name:         proto.Name,
		Slug:         proto.Slug,
		Description:  proto.Description,
		Type:         proto.Type,
		IsSystemRole: proto.IsSystemRole,
		Permissions:  proto.Permissions,
		Priority:     int(proto.Priority),
		Status:       proto.Status,
		CreatedBy:    proto.CreatedBy,
		CreatedAt:    time.Now(), // Set by service layer
		UpdatedAt:    time.Now(),
	}

	// Handle optional metadata
	if proto.Metadata != nil {
		role.Metadata = auth_models.RoleMetadata{
			Color:         proto.Metadata.Color,
			Icon:          proto.Metadata.Icon,
			MaxAssignable: int(proto.Metadata.MaxAssignable),
		}
	}

	return role, nil
}

// =============================================================================
// Proto → Domain Model (for update operations)
// =============================================================================

// UpdateRoleFromProto applies updates from gRPC UpdateRoleData to existing Role
// Only updates fields that are provided (non-nil for optional fields)
func UpdateRoleFromProto(existing *auth_models.Role, proto *authv1.UpdateRoleData) error {
	if existing == nil {
		return erp_errors.Validation(erp_errors.ValidationInvalidValue, "existing")
	}
	if proto == nil {
		return erp_errors.Validation(erp_errors.ValidationInvalidValue, "proto")
	}

	// Update only fields that are set (non-nil for optional fields in proto3)
	if proto.Name != nil {
		existing.Name = *proto.Name
	}
	if proto.Slug != nil {
		existing.Slug = *proto.Slug
	}
	if proto.Description != nil {
		existing.Description = *proto.Description
	}
	if proto.Type != nil {
		existing.Type = *proto.Type
	}
	if proto.IsSystemRole != nil {
		existing.IsSystemRole = *proto.IsSystemRole
	}

	// Permissions is a repeated field, if present replace the whole list
	if proto.Permissions != nil {
		existing.Permissions = proto.Permissions
	}

	if proto.Priority != nil {
		existing.Priority = int(*proto.Priority)
	}
	if proto.Status != nil {
		existing.Status = *proto.Status
	}
	if proto.Metadata != nil {
		existing.Metadata = auth_models.RoleMetadata{
			Color:         proto.Metadata.Color,
			Icon:          proto.Metadata.Icon,
			MaxAssignable: int(proto.Metadata.MaxAssignable),
		}
	}

	// Always update the timestamp
	existing.UpdatedAt = time.Now()

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// RoleObjectIDFromString converts hex string to primitive.ObjectID
func RoleObjectIDFromString(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, erp_errors.Validation(erp_errors.ValidationInvalidValue, "id")
	}
	return primitive.ObjectIDFromHex(id)
}
