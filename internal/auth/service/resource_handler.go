package service

import (
	"erp.localhost/internal/infra/convertor"
	erp_errors "erp.localhost/internal/infra/error"
	auth_models "erp.localhost/internal/infra/model/auth"
	auth_proto "erp.localhost/internal/infra/proto/auth/v1"
)

// resourceHandler knows how to handle a specific resource type
// This interface encapsulates all resource-specific conversion and extraction logic
type resourceHandler interface {
	// ExtractAndConvertCreate extracts the resource from the create request and converts to domain model
	ExtractAndConvertCreate(req *auth_proto.CreateResourceRequest) (any, error)

	// ExtractUpdateData extracts the update data from the update request
	ExtractUpdateData(req *auth_proto.UpdateResourceRequest) (any, error)

	// GetResourceIDFromUpdate extracts the resource ID from update data
	GetResourceIDFromUpdate(updateData any) string

	// ApplyUpdate applies the proto update data to the existing domain model
	ApplyUpdate(existing any, updateData any) error

	// GetResourceType returns the domain model resource type constant
	GetResourceType() string

	// GetResourceName returns a human-readable name for logging
	GetResourceName() string
}

// =============================================================================
// Permission Handler
// =============================================================================

// permissionHandler handles Permission resources
type permissionHandler struct{}

func (h *permissionHandler) ExtractAndConvertCreate(req *auth_proto.CreateResourceRequest) (any, error) {
	createResource, ok := req.GetResource().(*auth_proto.CreateResourceRequest_Permission)
	if !ok {
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidType, "resource")
	}

	permission, err := convertor.CreatePermissionFromProto(createResource.Permission)
	if err != nil {
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue, "permission").WithError(err)
	}

	return permission, nil
}

func (h *permissionHandler) ExtractUpdateData(req *auth_proto.UpdateResourceRequest) (any, error) {
	updateResource, ok := req.GetResource().(*auth_proto.UpdateResourceRequest_Permission)
	if !ok {
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidType, "resource")
	}
	return updateResource.Permission, nil
}

func (h *permissionHandler) GetResourceIDFromUpdate(updateData any) string {
	return updateData.(*auth_proto.UpdatePermissionData).GetId()
}

func (h *permissionHandler) ApplyUpdate(existing any, updateData any) error {
	permission, ok := existing.(*auth_models.Permission)
	if !ok {
		return erp_errors.Validation(erp_errors.ValidationInvalidType, "existing resource")
	}

	protoData, ok := updateData.(*auth_proto.UpdatePermissionData)
	if !ok {
		return erp_errors.Validation(erp_errors.ValidationInvalidType, "update data")
	}

	return convertor.UpdatePermissionFromProto(permission, protoData)
}

func (h *permissionHandler) GetResourceType() string {
	return auth_models.ResourceTypePermission
}

func (h *permissionHandler) GetResourceName() string {
	return "permission"
}

// =============================================================================
// Role Handler
// =============================================================================

// roleHandler handles Role resources
type roleHandler struct{}

func (h *roleHandler) ExtractAndConvertCreate(req *auth_proto.CreateResourceRequest) (any, error) {
	createResource, ok := req.GetResource().(*auth_proto.CreateResourceRequest_Role)
	if !ok {
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidType, "resource")
	}

	role, err := convertor.CreateRoleFromProto(createResource.Role)
	if err != nil {
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue, "role").WithError(err)
	}

	return role, nil
}

func (h *roleHandler) ExtractUpdateData(req *auth_proto.UpdateResourceRequest) (any, error) {
	updateResource, ok := req.GetResource().(*auth_proto.UpdateResourceRequest_Role)
	if !ok {
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidType, "resource")
	}
	return updateResource.Role, nil
}

func (h *roleHandler) GetResourceIDFromUpdate(updateData any) string {
	return updateData.(*auth_proto.UpdateRoleData).GetId()
}

func (h *roleHandler) ApplyUpdate(existing any, updateData any) error {
	role, ok := existing.(*auth_models.Role)
	if !ok {
		return erp_errors.Validation(erp_errors.ValidationInvalidType, "existing resource")
	}

	protoData, ok := updateData.(*auth_proto.UpdateRoleData)
	if !ok {
		return erp_errors.Validation(erp_errors.ValidationInvalidType, "update data")
	}

	return convertor.UpdateRoleFromProto(role, protoData)
}

func (h *roleHandler) GetResourceType() string {
	return auth_models.ResourceTypeRole
}

func (h *roleHandler) GetResourceName() string {
	return "role"
}

// =============================================================================
// Handler Registry
// =============================================================================

// resourceHandlers maps proto resource types to their handlers
var resourceHandlers = map[auth_proto.ResourceType]resourceHandler{
	auth_proto.ResourceType_RESOURCE_TYPE_PERMISSION: &permissionHandler{},
	auth_proto.ResourceType_RESOURCE_TYPE_ROLE:       &roleHandler{},
}

// getResourceHandler retrieves the appropriate handler for a given resource type
func getResourceHandler(resourceType auth_proto.ResourceType) (resourceHandler, error) {
	handler, ok := resourceHandlers[resourceType]
	if !ok {
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue, "resourceType")
	}
	return handler, nil
}
