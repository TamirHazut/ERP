package convertor

import (
	"strings"
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

// PermissionToProto converts internal Permission model to gRPC PermissionData
func PermissionToProto(perm *auth_models.Permission) *authv1.PermissionData {
	if perm == nil || perm.Validate(false) != nil {
		return nil
	}

	return &authv1.PermissionData{
		Id:          perm.ID.Hex(),
		TenantId:    perm.TenantID,
		Name:        perm.DisplayName,      // Map DisplayName to Name in proto
		Slug:        perm.PermissionString, // Map PermissionString to Slug in proto
		Description: perm.Description,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Status:      auth_models.PermissionStatusActive, // Default status (not in current model)
		CreatedAt:   timestamppb.New(perm.CreatedAt),
		UpdatedAt:   timestamppb.New(perm.UpdatedAt),
		CreatedBy:   perm.CreatedBy,
	}
}

// PermissionsToProto converts slice of permissions to proto
func PermissionsToProto(perms []auth_models.Permission) []*authv1.PermissionData {
	result := make([]*authv1.PermissionData, len(perms))
	for i, perm := range perms {
		result[i] = PermissionToProto(&perm)
	}
	return result
}

// =============================================================================
// Proto → Domain Model (for create operations)
// =============================================================================

// CreatePermissionFromProto converts gRPC CreatePermissionData to internal Permission
func CreatePermissionFromProto(proto *authv1.CreatePermissionData) (*auth_models.Permission, error) {
	if proto == nil {
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue, "proto")
	}

	perm := &auth_models.Permission{
		TenantID:         proto.TenantId,
		Resource:         proto.Resource,
		Action:           proto.Action,
		PermissionString: proto.Slug, // Map Slug to PermissionString
		DisplayName:      proto.Name, // Map Name to DisplayName
		Description:      proto.Description,
		Category:         "",         // Set default or extract from metadata
		IsDangerous:      false,      // Set default
		RequiresApproval: false,      // Set default
		Dependencies:     []string{}, // Initialize empty
		CreatedBy:        proto.CreatedBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata: auth_models.PermissionMetadata{
			Module:  "", // Could be derived from resource
			UIGroup: "", // Could be set based on category
		},
	}

	return perm, nil
}

// =============================================================================
// Proto → Domain Model (for update operations)
// =============================================================================

// UpdatePermissionFromProto applies updates from gRPC UpdatePermissionData to existing Permission
// Only updates fields that are provided (non-nil for optional fields)
func UpdatePermissionFromProto(existing *auth_models.Permission, proto *authv1.UpdatePermissionData) error {
	if existing == nil {
		return erp_errors.Validation(erp_errors.ValidationInvalidValue, "existing")
	}
	if proto == nil {
		return erp_errors.Validation(erp_errors.ValidationInvalidValue, "proto")
	}

	// Update only fields that are set (non-nil for optional fields in proto3)
	if proto.Name != nil {
		existing.DisplayName = *proto.Name
	}
	if proto.Slug != nil {
		existing.PermissionString = *proto.Slug
	}
	if proto.Description != nil {
		existing.Description = *proto.Description
	}
	if proto.Resource != nil {
		existing.Resource = *proto.Resource
	}
	if proto.Action != nil {
		existing.Action = *proto.Action
	}
	// Note: Status field doesn't exist in current Permission model
	// If added in the future, handle it here

	// Always update the timestamp
	existing.UpdatedAt = time.Now()

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// PermissionObjectIDFromString converts hex string to primitive.ObjectID
func PermissionObjectIDFromString(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, erp_errors.Validation(erp_errors.ValidationInvalidValue, "id")
	}
	return primitive.ObjectIDFromHex(id)
}

// =============================================================================
// Permission Identifier Converters (for verification)
// =============================================================================

// PermissionIdentifierToString converts permission identifier to a string format
// Format: "resource:action" (e.g., "users:create")
func PermissionIdentifierToString(identifier *authv1.PermissionIdentifier) string {
	if identifier == nil {
		return ""
	}
	return identifier.Resource + ":" + identifier.Action
}

// PermissionIdentifierFromString parses a permission string into components
// Input format: "resource:action" (e.g., "users:create")
func PermissionIdentifierFromString(permString string) *authv1.PermissionIdentifier {
	// This is a simple implementation - you may want to add validation
	// Parse the permission string (format: "resource:action")
	// For a more robust implementation, add proper parsing logic
	permSplt := strings.Split(permString, ":")
	if len(permSplt) != 2 {
		return nil
	}
	if !auth_models.IsValidResourceType(permSplt[0]) || !auth_models.IsValidPermissionAction(permSplt[1]) {
		return nil
	}
	return &authv1.PermissionIdentifier{
		Resource: strings.ToLower(permSplt[0]), // Parse from permString
		Action:   strings.ToLower(permSplt[1]), // Parse from permString
	}
}
