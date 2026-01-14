package convertor

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	infra_error "erp.localhost/internal/infra/error"
	model_auth "erp.localhost/internal/infra/model/auth"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
)

// =============================================================================
// Domain Model → Proto (for responses)
// =============================================================================

// RoleToProto converts internal Role model to gRPC RoleData
func RoleToProto(role *model_auth.Role) *proto_auth.RoleData {
	if role == nil || role.Validate(false) != nil {
		return nil
	}

	return &proto_auth.RoleData{
		Id:            role.ID.Hex(),
		TenantId:      role.TenantID,
		Name:          role.Name,
		Slug:          role.Slug,
		Description:   role.Description,
		Type:          role.Type,
		IsTenantAdmin: role.IsTenantAdmin,
		Permissions:   role.Permissions,
		Priority:      int32(role.Priority),
		Status:        role.Status,
		Metadata:      RoleMetadataToProto(&role.Metadata),
		CreatedAt:     timestamppb.New(role.CreatedAt),
		UpdatedAt:     timestamppb.New(role.UpdatedAt),
		CreatedBy:     role.CreatedBy,
	}
}

// RoleMetadataToProto converts metadata
func RoleMetadataToProto(metadata *model_auth.RoleMetadata) *proto_auth.RoleMetadata {
	if metadata == nil {
		return nil
	}

	return &proto_auth.RoleMetadata{
		Color:         metadata.Color,
		Icon:          metadata.Icon,
		MaxAssignable: int32(metadata.MaxAssignable),
	}
}

// RolesToProto converts slice of roles to proto
func RolesToProto(roles []model_auth.Role) []*proto_auth.RoleData {
	result := make([]*proto_auth.RoleData, len(roles))
	for i, role := range roles {
		result[i] = RoleToProto(&role)
	}
	return result
}

func ProtoToRole(pb *proto_auth.RoleData) (*model_auth.Role, error) {
	if pb == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "role")
	}
	id, _ := primitive.ObjectIDFromHex(pb.Id)

	return &model_auth.Role{
		ID:            id,
		TenantID:      pb.TenantId,
		Name:          pb.Name,
		Slug:          pb.Slug,
		Description:   pb.Description,
		Type:          pb.Type,
		IsTenantAdmin: pb.IsTenantAdmin,
		Permissions:   pb.Permissions,
		Priority:      int(pb.Priority),
		Status:        pb.Status,
		Metadata: model_auth.RoleMetadata{
			Color:         pb.Metadata.GetColor(),
			Icon:          pb.Metadata.GetIcon(),
			MaxAssignable: int(pb.Metadata.GetMaxAssignable()),
		},
		CreatedAt: pb.CreatedAt.AsTime(),
		UpdatedAt: pb.UpdatedAt.AsTime(),
		CreatedBy: pb.CreatedBy,
	}, nil
}

// =============================================================================
// Proto → Domain Model (for create operations)
// =============================================================================

// CreateRoleFromProto converts gRPC CreateRoleData to internal Role model
func CreateRoleFromProto(proto *proto_auth.CreateRoleData) (*model_auth.Role, error) {
	if proto == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "proto")
	}

	role := &model_auth.Role{
		TenantID:      proto.TenantId,
		Name:          proto.Name,
		Slug:          proto.Slug,
		Description:   proto.Description,
		Type:          proto.Type,
		IsTenantAdmin: proto.IsTenantAdmin,
		Permissions:   proto.Permissions,
		Priority:      int(proto.Priority),
		Status:        proto.Status,
		CreatedBy:     proto.CreatedBy,
		CreatedAt:     time.Now(), // Set by service layer
		UpdatedAt:     time.Now(),
	}

	// Handle optional metadata
	if proto.Metadata != nil {
		role.Metadata = model_auth.RoleMetadata{
			Color:         proto.Metadata.Color,
			Icon:          proto.Metadata.Icon,
			MaxAssignable: int(proto.Metadata.MaxAssignable),
		}
	}

	return role, nil
}

func RoleToCreateProto(role *model_auth.Role) (*proto_auth.CreateRoleData, error) {
	if role == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "role")
	}
	return &proto_auth.CreateRoleData{
		TenantId:      role.TenantID,
		Name:          role.Name,
		Slug:          role.Slug,
		Description:   role.Description,
		Type:          role.Type,
		IsTenantAdmin: role.IsTenantAdmin,
		Permissions:   role.Permissions,
		Priority:      int32(role.Priority),
		Status:        role.Status,
		Metadata: &proto_auth.RoleMetadata{
			Color:         role.Metadata.Color,
			Icon:          role.Metadata.Icon,
			MaxAssignable: int32(role.Metadata.MaxAssignable),
		},
		CreatedBy: role.CreatedBy,
	}, nil
}

// =============================================================================
// Proto → Domain Model (for update operations)
// =============================================================================

// UpdateRoleFromProto applies updates from gRPC UpdateRoleData to existing Role
// Only updates fields that are provided (non-nil for optional fields)
func UpdateRoleFromProto(existing *model_auth.Role, proto *proto_auth.UpdateRoleData) error {
	if existing == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "existing")
	}
	if proto == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "proto")
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
	if proto.IsTenantAdmin != nil {
		existing.IsTenantAdmin = *proto.IsTenantAdmin
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
		existing.Metadata = model_auth.RoleMetadata{
			Color:         proto.Metadata.Color,
			Icon:          proto.Metadata.Icon,
			MaxAssignable: int(proto.Metadata.MaxAssignable),
		}
	}

	// Always update the timestamp
	existing.UpdatedAt = time.Now()

	return nil
}

func RoleToUpdateProto(role *model_auth.Role) (*proto_auth.UpdateRoleData, error) {
	if role == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "role")
	}
	return &proto_auth.UpdateRoleData{
		Id:            role.ID.Hex(),
		TenantId:      role.TenantID,
		Name:          &role.Name,
		Slug:          &role.Slug,
		Description:   &role.Description,
		Type:          &role.Type,
		IsTenantAdmin: &role.IsTenantAdmin,
		Permissions:   role.Permissions,
		Priority:      ptrInt32(int32(role.Priority)),
		Status:        &role.Status,
		Metadata: &proto_auth.RoleMetadata{
			Color:         role.Metadata.Color,
			Icon:          role.Metadata.Icon,
			MaxAssignable: int32(role.Metadata.MaxAssignable),
		},
	}, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// RoleObjectIDFromString converts hex string to primitive.ObjectID
func RoleObjectIDFromString(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, infra_error.Validation(infra_error.ValidationInvalidValue, "id")
	}
	return primitive.ObjectIDFromHex(id)
}

// ============================================================================
// Helper functions
// ============================================================================

func ptrInt32(v int32) *int32 {
	return &v
}
