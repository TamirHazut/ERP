package rbac

import (
	"fmt"
	"slices"

	collection "erp.localhost/internal/auth/collections"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
	shared_models "erp.localhost/internal/shared/models"
	auth_models "erp.localhost/internal/shared/models/auth"
)

type RBACManager struct {
	logger                *logging.Logger
	rolesCollection       *collection.RolesCollection
	permissionsCollection *collection.PermissionsCollection
}

func NewRBACManager() *RBACManager {
	logger := logging.NewLogger(shared_models.ModuleAuth)
	return &RBACManager{
		logger:                logger,
		rolesCollection:       collection.NewRoleCollection(nil),
		permissionsCollection: collection.NewPermissionCollection(nil),
	}
}

/* CRUD Resource Functions */
func (r *RBACManager) CreateResource(tenantID string, userID string, resourceType string, resource any) error {
	permission := fmt.Sprintf("%s:%s", resourceType, auth_models.PermissionActionCreate)

	return r.HasPermission(tenantID, userID, permission)
}

func (r *RBACManager) UpdateResource(tenantID string, userID string, resourceType string, resource any) error {
	permission := fmt.Sprintf("%s:%s", resourceType, auth_models.PermissionActionUpdate)

	return r.HasPermission(tenantID, userID, permission)
}

func (r *RBACManager) DeleteResource(tenantID string, userID string, resourceType string, resource any) error {
	permission := fmt.Sprintf("%s:%s", resourceType, auth_models.PermissionActionDelete)

	return r.HasPermission(tenantID, userID, permission)
}

func (r *RBACManager) GetResource(tenantID string, userID string, resourceType string, resourceID string) error {
	permission := fmt.Sprintf("%s:%s", resourceType, auth_models.PermissionActionRead)

	return r.HasPermission(tenantID, userID, permission)
}

func (r *RBACManager) GetResources(tenantID string, userID string, resourceType string) error {
	permission := fmt.Sprintf("%s:%s", resourceType, auth_models.PermissionActionRead)

	return r.HasPermission(tenantID, userID, permission)
}

/* Get Functions */
// TODO: uncomment when user grpc service is ready
func (r *RBACManager) GetUserPermissions(tenantID string, userID string) (map[string]bool, error) {
	/*user, err := r.userCollection.GetUserByID(tenantID, userID)
	if err != nil {
		return nil, err
	}
	userRoles := user.Roles
	userPermissions := make(map[string]bool, 0)
	for _, userRole := range userRoles {
		role, err := r.rolesCollection.GetRoleByID(tenantID, userRole.RoleID)
		if err != nil {
			return nil, err
		}
		for _, permission := range role.Permissions {
			userPermissions[permission] = true
		}
	}
	for _, permission := range user.AdditionalPermissions {
		userPermissions[permission] = true
	}
	for _, permission := range user.RevokedPermissions {
		userPermissions[permission] = false
	}
	return userPermissions, nil*/
	return map[string]bool{}, nil
}

// TODO: uncomment when user grpc service is ready
func (r *RBACManager) GetUserRoles(tenantID string, userID string) ([]string, error) {
	// user, err := r.userCollection.GetUserByID(tenantID, userID)
	// if err != nil {
	// 	return nil, err
	// }
	// userRoles := make([]string, 0)
	// for _, userRole := range user.Roles {
	// 	userRoles = append(userRoles, userRole.RoleID)
	// }
	// return userRoles, nil
	return []string{}, nil
}

func (r *RBACManager) GetRolePermissions(tenantID string, roleID string) ([]string, error) {
	role, err := r.rolesCollection.GetRoleByID(tenantID, roleID)
	if err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

/* Verification Functions */
func (r *RBACManager) CheckUserPermissions(tenantID string, userID string, permissions []string) (map[string]bool, error) {
	if tenantID == "" || userID == "" || len(permissions) == 0 {
		return nil, erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenantID", "userID", "permission")
	}

	// Verify permissions format
	for _, permission := range permissions {
		if !auth_models.IsValidPermissionFormat(permission) {
			return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue).WithError(fmt.Errorf("invalid permission format"))
		}
	}

	// Get user permissions
	userPermissions, err := r.GetUserPermissions(tenantID, userID)
	if err != nil {
		return nil, err
	}
	// Check requested permissions status
	permissionsCheckResponse := make(map[string]bool, 0)
	for _, permission := range permissions {
		if val, ok := userPermissions[permission]; ok {
			permissionsCheckResponse[permission] = val
		} else {
			permissionsCheckResponse[permission] = false
		}
	}
	return permissionsCheckResponse, nil
}

func (r *RBACManager) VerifyUserRole(tenantID string, userID string, roleID string) (bool, error) {
	userRoles, err := r.GetUserRoles(tenantID, userID)
	if err != nil {
		return false, err
	}
	for _, userRole := range userRoles {
		if userRole == roleID {
			return true, nil
		}
	}
	return false, nil
}

func (r *RBACManager) VerifyRolePermissions(tenantID string, roleID string, permissions []string) (map[string]bool, error) {
	rolePermissions, err := r.GetRolePermissions(tenantID, roleID)
	if err != nil {
		return nil, err
	}
	permissionsCheckResponse := make(map[string]bool, 0)
	for _, permission := range permissions {
		permissionsCheckResponse[permission] = slices.Contains(rolePermissions, permission)
	}
	return permissionsCheckResponse, nil
}

func (r *RBACManager) HasPermission(tenantID string, userID string, permission string) error {
	permissionsCheckResponse, err := r.CheckUserPermissions(tenantID, userID, []string{permission})
	if err != nil {
		return err
	}
	if !permissionsCheckResponse[permission] {
		return erp_errors.Auth(erp_errors.AuthPermissionDenied)
	}
	return nil
}
