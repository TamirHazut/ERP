package convertor

import (
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	infra_error "erp.localhost/internal/infra/error"
	model_auth "erp.localhost/internal/infra/model/auth"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
)

// =============================================================================
// Domain Model → Proto (for responses)
// =============================================================================

// TenantToProto converts a Tenant model to TenantData proto message
func TenantToProto(tenant *model_auth.Tenant) *proto_auth.TenantData {
	if tenant == nil || tenant.Validate(false) != nil {
		return nil
	}

	return &proto_auth.TenantData{
		Id:        tenant.ID.Hex(),
		Name:      tenant.Name,
		Status:    tenant.Status,
		CreatedAt: timestamppb.New(tenant.CreatedAt),
		UpdatedAt: timestamppb.New(tenant.UpdatedAt),
		CreatedBy: tenant.CreatedBy,
	}
}

// TenantsToProto converts a slice of Tenant models to TenantData proto messages
func TenantsToProto(tenants []*model_auth.Tenant) []*proto_auth.TenantData {
	if tenants == nil {
		return []*proto_auth.TenantData{}
	}

	protoTenants := make([]*proto_auth.TenantData, 0, len(tenants))
	for _, tenant := range tenants {
		if protoTenant := TenantToProto(tenant); protoTenant != nil {
			protoTenants = append(protoTenants, protoTenant)
		}
	}
	return protoTenants
}

// =============================================================================
// Proto → Domain Model (for create operations)
// =============================================================================

// CreateTenantFromProto converts a CreateTenantData proto message to a Tenant model
func CreateTenantFromProto(proto *proto_auth.CreateTenantData) (*model_auth.Tenant, error) {
	if proto == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "proto")
	}

	// Validate required fields
	if proto.Name == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "name")
	}
	if proto.CreatedBy == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "created_by")
	}

	// Generate slug from name (lowercase, spaces → dashes)
	slug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(proto.Name), " ", "-"))

	// Set default status if not provided
	status := proto.Status
	if status == "" {
		status = "active"
	}

	now := time.Now()

	tenant := &model_auth.Tenant{
		Name:      proto.Name,
		Slug:      slug,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: proto.CreatedBy,
		// Initialize nested structs with zero values
		Subscription: model_auth.Subscription{},
		Settings:     model_auth.TenantSettings{},
		Contact:      model_auth.ContactInfo{},
		Branding:     model_auth.Branding{},
		Metadata:     model_auth.TenantMetadata{},
	}

	return tenant, nil
}

// =============================================================================
// Proto → Domain Model (for update operations)
// =============================================================================

// UpdateTenantFromProto applies updates from UpdateTenantData proto to an existing Tenant model
func UpdateTenantFromProto(existing *model_auth.Tenant, proto *proto_auth.UpdateTenantData) error {
	if existing == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "existing")
	}
	if proto == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "proto")
	}

	// Update simple fields if provided
	if proto.Name != nil {
		existing.Name = *proto.Name
		// Update slug when name changes
		existing.Slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(*proto.Name), " ", "-"))
	}
	if proto.Status != nil {
		existing.Status = *proto.Status
	}

	// Always update timestamp
	existing.UpdatedAt = time.Now()

	return nil
}
