package handler

import (
	"fmt"
	"strings"

	"erp.localhost/internal/auth/collection"
	infra_error "erp.localhost/internal/infra/error"
	model_auth "erp.localhost/internal/infra/model/auth"
)

//go:generate mockgen -destination=mock/mock_rbac_handler.go -package=mock erp.localhost/internal/auth/rbac/handler RBACResourceHandler

// RBACHandler is for the CRUD operations of RBAC types (Roles, Permissions)
type RBACResourceHandler interface {
	CreateResource(resource model_auth.RBACResource) (string, error)
	UpdateResource(resource model_auth.RBACResource) error
	DeleteResource(tenantID string, userID string) error
	GetResource(tenantID string, filter SearchFilter) (model_auth.RBACResource, error)
	ListResources(tenantID string, filter SearchFilter) ([]model_auth.RBACResource, error)
}

func NewPermissionHandler(collection *collection.PermissionsCollection) *BaseRBACResourceHandler[*model_auth.Permission] {
	return &BaseRBACResourceHandler[*model_auth.Permission]{
		resourceType: model_auth.ResourceTypePermission,
		createFunc: func(resource *model_auth.Permission) (string, error) {
			return collection.CreatePermission(resource)
		},
		updateFunc: func(resource *model_auth.Permission) error {
			return collection.UpdatePermission(resource)
		},
		deleteFunc: func(tenantID string, resourceID string) error {
			return collection.DeletePermission(tenantID, resourceID)
		},
		getOneFunc: func(tenantID string, filter SearchFilter) (*model_auth.Permission, error) {
			if filter != nil {
				if resourceID, ok := filter[FilterKeyResourceID]; ok {
					return collection.GetPermissionByID(tenantID, resourceID)
				}
				if name, ok := filter[FilterKeyName]; ok {
					return collection.GetPermissionByID(tenantID, name)
				}
			}
			return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "filter")
		},
		getAllFunc: func(tenantID string, filter SearchFilter) ([]*model_auth.Permission, error) {
			if filter == nil {
				return collection.GetPermissionsByTenantID(tenantID)
			}
			resourceType, resourceTypeOk := filter[FilterKeyResourceType]
			action, actionOk := filter[FilterKeyAction]
			if resourceTypeOk && actionOk {
				return collection.GetPermissionsByResourceAndAction(tenantID, resourceType, action)
			} else if resourceTypeOk {
				return collection.GetPermissionsByResource(tenantID, resourceType)
			} else if actionOk {
				return collection.GetPermissionsByAction(tenantID, action)
			}
			return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "filter")
		},
	}
}

func NewRoleHandler(collection *collection.RolesCollection) *BaseRBACResourceHandler[*model_auth.Role] {
	return &BaseRBACResourceHandler[*model_auth.Role]{
		resourceType: model_auth.ResourceTypeRole,
		createFunc: func(resource *model_auth.Role) (string, error) {
			return collection.CreateRole(resource)
		},
		updateFunc: func(resource *model_auth.Role) error {
			return collection.UpdateRole(resource)
		},
		deleteFunc: func(tenantID string, resourceID string) error {
			return collection.DeleteRole(tenantID, resourceID)
		},
		getOneFunc: func(tenantID string, filter SearchFilter) (*model_auth.Role, error) {
			if filter != nil {
				if resourceID, ok := filter[FilterKeyResourceID]; ok {
					return collection.GetRoleByID(tenantID, resourceID)
				}
				if name, ok := filter[FilterKeyName]; ok {
					return collection.GetRoleByName(tenantID, name)
				}
			}
			return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "filter")
		},
		getAllFunc: func(tenantID string, filter SearchFilter) ([]*model_auth.Role, error) {
			if filter == nil {
				return collection.GetRolesByTenantID(tenantID)
			}
			if permissionID, ok := filter[FilterKeyPermissionsID]; ok {
				permissions := strings.Split(permissionID, ",")
				return collection.GetRolesByPermissionsIDs(tenantID, permissions)
			}
			return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "filter")
		},
	}
}

type SearchFilter map[FilterKey]string

type FilterKey string

const (
	FilterKeyTenantID      FilterKey = "tenant_id"
	FilterKeyResourceID    FilterKey = "resource_id"
	FilterKeyUserID        FilterKey = "user_id"
	FilterKeyPermissionsID FilterKey = "permissions_id" // comma separated
	FilterKeyName          FilterKey = "name"
	FilterKeyResourceType  FilterKey = "resource_type"
	FilterKeyAction        FilterKey = "action"
)

type BaseRBACResourceHandler[T model_auth.RBACResource] struct {
	resourceType string
	createFunc   func(resource T) (string, error)
	updateFunc   func(resource T) error
	deleteFunc   func(tenantID string, resourceID string) error
	getOneFunc   func(tenantID string, filter SearchFilter) (T, error)
	getAllFunc   func(tenantID string, filter SearchFilter) ([]T, error)
}

func (h *BaseRBACResourceHandler[T]) CreateResource(resource model_auth.RBACResource) (string, error) {
	typedResource, ok := resource.(T)
	if !ok {
		return "", infra_error.Validation(infra_error.ValidationInvalidValue).
			WithError(fmt.Errorf("invalid resource type: expected %s, got %T", h.resourceType, resource))
	}
	// Validate resource type matches handler type
	if typedResource.GetResourceType() != h.resourceType {
		return "", infra_error.Validation(infra_error.ValidationInvalidValue).
			WithError(fmt.Errorf("resource type mismatch: expected %s, got %s",
				h.resourceType, typedResource.GetResourceType()))
	}

	// Delegate to the type-specific creation function
	return h.createFunc(typedResource)
}

func (h *BaseRBACResourceHandler[T]) UpdateResource(resource model_auth.RBACResource) error {
	typedResource, ok := resource.(T)
	if !ok {
		return infra_error.Validation(infra_error.ValidationInvalidValue).
			WithError(fmt.Errorf("invalid resource type: expected %s, got %T", h.resourceType, resource))
	}
	// Validate resource type matches handler type
	if typedResource.GetResourceType() != h.resourceType {
		return infra_error.Validation(infra_error.ValidationInvalidValue).
			WithError(fmt.Errorf("resource type mismatch: expected %s, got %s",
				h.resourceType, typedResource.GetResourceType()))
	}
	return h.updateFunc(typedResource)
}

func (h *BaseRBACResourceHandler[T]) DeleteResource(tenantID string, resourceID string) error {
	return h.deleteFunc(tenantID, resourceID)
}

func (h *BaseRBACResourceHandler[T]) GetResource(tenantID string, filter SearchFilter) (model_auth.RBACResource, error) {
	return h.getOneFunc(tenantID, filter)
}

func (h *BaseRBACResourceHandler[T]) ListResources(tenantID string, filter SearchFilter) ([]model_auth.RBACResource, error) {
	resources, err := h.getAllFunc(tenantID, filter)
	if err != nil {
		return nil, err
	}

	// Convert []T to []RBACResource
	result := make([]model_auth.RBACResource, len(resources))
	for i, r := range resources {
		result[i] = r
	}
	return result, nil
}
