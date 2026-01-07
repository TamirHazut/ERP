package service

import (
	"erp.localhost/internal/infra/convertor"
	error_infra "erp.localhost/internal/infra/error"
	model_auth "erp.localhost/internal/infra/model/auth"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
)

// resourceHandler knows how to handle a specific resource type
// This interface encapsulates all resource-specific conversion and extraction logic
type resourceHandler interface {
	// ExtractAndConvertCreate extracts the resource from the create request and converts to domain model
	ExtractAndConvertCreate(req *proto_auth.CreateResourceRequest) (model_auth.RBACResource, error)

	// ExtractUpdateData extracts the update data from the update request
	ExtractUpdateData(req *proto_auth.UpdateResourceRequest) (any, error)

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

func (h *permissionHandler) ExtractAndConvertCreate(req *proto_auth.CreateResourceRequest) (model_auth.RBACResource, error) {
	createResource, ok := req.GetResource().(*proto_auth.CreateResourceRequest_Permission)
	if !ok {
		return nil, error_infra.Validation(error_infra.ValidationInvalidType, "resource")
	}

	permission, err := convertor.CreatePermissionFromProto(createResource.Permission)
	if err != nil {
		return nil, error_infra.Validation(error_infra.ValidationInvalidValue, "permission").WithError(err)
	}

	return permission, nil
}

func (h *permissionHandler) ExtractUpdateData(req *proto_auth.UpdateResourceRequest) (any, error) {
	updateResource, ok := req.GetResource().(*proto_auth.UpdateResourceRequest_Permission)
	if !ok {
		return nil, error_infra.Validation(error_infra.ValidationInvalidType, "resource")
	}
	return updateResource.Permission, nil
}

func (h *permissionHandler) GetResourceIDFromUpdate(updateData any) string {
	return updateData.(*proto_auth.UpdatePermissionData).GetId()
}

func (h *permissionHandler) ApplyUpdate(existing any, updateData any) error {
	permission, ok := existing.(*model_auth.Permission)
	if !ok {
		return error_infra.Validation(error_infra.ValidationInvalidType, "existing resource")
	}

	protoData, ok := updateData.(*proto_auth.UpdatePermissionData)
	if !ok {
		return error_infra.Validation(error_infra.ValidationInvalidType, "update data")
	}

	return convertor.UpdatePermissionFromProto(permission, protoData)
}

func (h *permissionHandler) GetResourceType() string {
	return model_auth.ResourceTypePermission
}

func (h *permissionHandler) GetResourceName() string {
	return "permission"
}

// =============================================================================
// Role Handler
// =============================================================================

// roleHandler handles Role resources
type roleHandler struct{}

func (h *roleHandler) ExtractAndConvertCreate(req *proto_auth.CreateResourceRequest) (model_auth.RBACResource, error) {
	createResource, ok := req.GetResource().(*proto_auth.CreateResourceRequest_Role)
	if !ok {
		return nil, error_infra.Validation(error_infra.ValidationInvalidType, "resource")
	}

	role, err := convertor.CreateRoleFromProto(createResource.Role)
	if err != nil {
		return nil, error_infra.Validation(error_infra.ValidationInvalidValue, "role").WithError(err)
	}

	return role, nil
}

func (h *roleHandler) ExtractUpdateData(req *proto_auth.UpdateResourceRequest) (any, error) {
	updateResource, ok := req.GetResource().(*proto_auth.UpdateResourceRequest_Role)
	if !ok {
		return nil, error_infra.Validation(error_infra.ValidationInvalidType, "resource")
	}
	return updateResource.Role, nil
}

func (h *roleHandler) GetResourceIDFromUpdate(updateData any) string {
	return updateData.(*proto_auth.UpdateRoleData).GetId()
}

func (h *roleHandler) ApplyUpdate(existing any, updateData any) error {
	role, ok := existing.(*model_auth.Role)
	if !ok {
		return error_infra.Validation(error_infra.ValidationInvalidType, "existing resource")
	}

	protoData, ok := updateData.(*proto_auth.UpdateRoleData)
	if !ok {
		return error_infra.Validation(error_infra.ValidationInvalidType, "update data")
	}

	return convertor.UpdateRoleFromProto(role, protoData)
}

func (h *roleHandler) GetResourceType() string {
	return model_auth.ResourceTypeRole
}

func (h *roleHandler) GetResourceName() string {
	return "role"
}

// =============================================================================
// Handler Registry
// =============================================================================

// resourceHandlers maps proto resource types to their handlers
var resourceHandlers = map[proto_auth.ResourceType]resourceHandler{
	proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION: &permissionHandler{},
	proto_auth.ResourceType_RESOURCE_TYPE_ROLE:       &roleHandler{},
}

// getResourceHandler retrieves the appropriate handler for a given resource type
func getResourceHandler(resourceType proto_auth.ResourceType) (resourceHandler, error) {
	handler, ok := resourceHandlers[resourceType]
	if !ok {
		return nil, error_infra.Validation(error_infra.ValidationInvalidValue, "resourceType")
	}
	return handler, nil
}
