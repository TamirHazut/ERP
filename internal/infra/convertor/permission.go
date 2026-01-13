package convertor

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	infra_error "erp.localhost/internal/infra/error"
	model_auth "erp.localhost/internal/infra/model/auth"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
)

// =============================================================================
// Domain Model → Proto (for responses)
// =============================================================================

// PermissionToProto converts internal Permission model to gRPC PermissionData
func PermissionToProto(perm *model_auth.Permission) *proto_auth.PermissionData {
	if perm == nil || perm.Validate(false) != nil {
		return nil
	}

	return &proto_auth.PermissionData{
		Id:          perm.ID.Hex(),
		TenantId:    perm.TenantID,
		Name:        perm.DisplayName,      // Map DisplayName to Name in proto
		Slug:        perm.PermissionString, // Map PermissionString to Slug in proto
		Description: perm.Description,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Status:      model_auth.PermissionStatusActive, // Default status (not in current model)
		CreatedAt:   timestamppb.New(perm.CreatedAt),
		UpdatedAt:   timestamppb.New(perm.UpdatedAt),
		CreatedBy:   perm.CreatedBy,
	}
}

// PermissionsToProto converts slice of permissions to proto
func PermissionsToProto(perms []model_auth.Permission) []*proto_auth.PermissionData {
	result := make([]*proto_auth.PermissionData, len(perms))
	for i, perm := range perms {
		result[i] = PermissionToProto(&perm)
	}
	return result
}

func ProtoToPermission(pb *proto_auth.PermissionData) (*model_auth.Permission, error) {
	if pb == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "permission")
	}
	id, _ := primitive.ObjectIDFromHex(pb.Id)

	return &model_auth.Permission{
		ID:               id,
		TenantID:         pb.TenantId,
		Resource:         pb.Resource,
		Action:           pb.Action,
		PermissionString: pb.Slug,
		DisplayName:      pb.Name,
		Description:      pb.Description,
		CreatedAt:        pb.CreatedAt.AsTime(),
		UpdatedAt:        pb.UpdatedAt.AsTime(),
		CreatedBy:        pb.CreatedBy,
	}, nil
}

// =============================================================================
// Proto → Domain Model (for create operations)
// =============================================================================

// CreatePermissionFromProto converts gRPC CreatePermissionData to internal Permission
func CreatePermissionFromProto(proto *proto_auth.CreatePermissionData) (*model_auth.Permission, error) {
	if proto == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "proto")
	}

	perm := &model_auth.Permission{
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
		Metadata: model_auth.PermissionMetadata{
			Module:  "", // Could be derived from resource
			UIGroup: "", // Could be set based on category
		},
	}

	return perm, nil
}

func PermissionToCreateProto(perm *model_auth.Permission) (*proto_auth.CreatePermissionData, error) {
	if perm == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "permission")
	}
	return &proto_auth.CreatePermissionData{
		TenantId:    perm.TenantID,
		Name:        perm.DisplayName,
		Slug:        perm.PermissionString,
		Description: perm.Description,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Status:      "active", // Default status
		CreatedBy:   perm.CreatedBy,
	}, nil
}

// =============================================================================
// Proto → Domain Model (for update operations)
// =============================================================================

// UpdatePermissionFromProto applies updates from gRPC UpdatePermissionData to existing Permission
// Only updates fields that are provided (non-nil for optional fields)
func UpdatePermissionFromProto(existing *model_auth.Permission, proto *proto_auth.UpdatePermissionData) error {
	if existing == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "existing")
	}
	if proto == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "proto")
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

func PermissionToUpdateProto(perm *model_auth.Permission) (*proto_auth.UpdatePermissionData, error) {
	if perm == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "permission")
	}
	return &proto_auth.UpdatePermissionData{
		Id:          perm.ID.Hex(),
		TenantId:    perm.TenantID,
		Name:        &perm.DisplayName,
		Slug:        &perm.PermissionString,
		Description: &perm.Description,
		Resource:    &perm.Resource,
		Action:      &perm.Action,
	}, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// PermissionObjectIDFromString converts hex string to primitive.ObjectID
func PermissionObjectIDFromString(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, infra_error.Validation(infra_error.ValidationInvalidValue, "id")
	}
	return primitive.ObjectIDFromHex(id)
}

// =============================================================================
// Permission Identifier Converters (for verification)
// =============================================================================

// PermissionIdentifierToString converts permission identifier to a string format
// Format: "resource:action" (e.g., "users:create")
func PermissionIdentifierToString(identifier *proto_auth.PermissionIdentifier) string {
	if identifier == nil {
		return ""
	}
	return identifier.Resource + ":" + identifier.Action
}

// PermissionIdentifierFromString parses a permission string into components
// Input format: "resource:action" (e.g., "users:create")
func PermissionIdentifierFromString(permString string) *proto_auth.PermissionIdentifier {
	// This is a simple implementation - you may want to add validation
	// Parse the permission string (format: "resource:action")
	// For a more robust implementation, add proper parsing logic
	permSplt := strings.Split(permString, ":")
	if len(permSplt) != 2 {
		return nil
	}
	if !model_auth.IsValidResourceType(permSplt[0]) || !model_auth.IsValidPermissionAction(permSplt[1]) {
		return nil
	}
	return &proto_auth.PermissionIdentifier{
		Resource: strings.ToLower(permSplt[0]), // Parse from permString
		Action:   strings.ToLower(permSplt[1]), // Parse from permString
	}
}
