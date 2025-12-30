package rbac

import (
	"fmt"
	"slices"

	collection "erp.localhost/internal/auth/collections"
	"erp.localhost/internal/auth/models"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

type RBACManager struct {
	logger                *logging.Logger
	userCollection        *collection.UserCollection
	rolesCollection       *collection.RolesCollection
	permissionsCollection *collection.PermissionsCollection
	auditLogsCollection   *collection.AuditLogsCollection
}

func NewRBACManager() *RBACManager {
	logger := logging.NewLogger(logging.ModuleAuth)
	return &RBACManager{
		logger:                logger,
		userCollection:        collection.NewUserCollection(nil),
		rolesCollection:       collection.NewRoleCollection(nil),
		permissionsCollection: collection.NewPermissionCollection(nil),
		auditLogsCollection:   collection.NewAuditLogsCollection(nil),
	}
}

/* CRUD Resource Functions */
func (r *RBACManager) CreateResource(tenantID string, userID string, resourceType string, resource any) (string, error) {
	permission := fmt.Sprintf("%s:%s", resourceType, models.PermissionActionCreate)

	if err := r.hasPermission(tenantID, userID, permission); err != nil {
		return "", err
	}

	invalidResourceValueTypeError := erp_errors.Validation(erp_errors.ValidationInvalidValue).WithError(fmt.Errorf("invalid resource value type"))
	switch resourceType {
	case models.ResourceTypeUser:
		user, ok := resource.(models.User)
		if !ok {
			return "", invalidResourceValueTypeError
		}
		return r.userCollection.CreateUser(user)
	case models.ResourceTypeRole:
		role, ok := resource.(models.Role)
		if !ok {
			return "", invalidResourceValueTypeError
		}
		return r.rolesCollection.CreateRole(role)
	case models.ResourceTypePermission:
		permission, ok := resource.(models.Permission)
		if !ok {
			return "", invalidResourceValueTypeError
		}
		return r.permissionsCollection.CreatePermission(permission)
	default:
		return "", erp_errors.Validation(erp_errors.ValidationInvalidValue).WithError(fmt.Errorf("unsupported resource type"))
	}
}

func (r *RBACManager) UpdateResource(tenantID string, userID string, resourceType string, resource any) error {
	permission := fmt.Sprintf("%s:%s", resourceType, models.PermissionActionUpdate)

	if err := r.hasPermission(tenantID, userID, permission); err != nil {
		return err
	}

	invalidResourceValueTypeError := erp_errors.Validation(erp_errors.ValidationInvalidValue).WithError(fmt.Errorf("invalid resource value type"))
	switch resourceType {
	case models.ResourceTypeUser:
		user, ok := resource.(models.User)
		if !ok {
			return invalidResourceValueTypeError
		}
		return r.userCollection.UpdateUser(user)
	case models.ResourceTypeRole:
		role, ok := resource.(models.Role)
		if !ok {
			return invalidResourceValueTypeError
		}
		return r.rolesCollection.UpdateRole(role)
	case models.ResourceTypePermission:
		permission, ok := resource.(models.Permission)
		if !ok {
			return invalidResourceValueTypeError
		}
		return r.permissionsCollection.UpdatePermission(permission)
	default:
		return erp_errors.Validation(erp_errors.ValidationInvalidValue).WithError(fmt.Errorf("unsupported resource type"))
	}
}

func (r *RBACManager) DeleteResource(tenantID string, userID string, resourceType string, resource any) error {
	permission := fmt.Sprintf("%s:%s", resourceType, models.PermissionActionDelete)
	if err := r.hasPermission(tenantID, userID, permission); err != nil {
		return err
	}
	invalidResourceValueTypeError := erp_errors.Validation(erp_errors.ValidationInvalidValue).WithError(fmt.Errorf("invalid resource value type"))
	switch resourceType {
	case models.ResourceTypeUser:
		user, ok := resource.(models.User)
		if !ok {
			return invalidResourceValueTypeError
		}
		return r.userCollection.DeleteUser(tenantID, user.ID.String())
	case models.ResourceTypeRole:
		role, ok := resource.(models.Role)
		if !ok {
			return invalidResourceValueTypeError
		}
		return r.rolesCollection.DeleteRole(tenantID, role.ID.String())
	case models.ResourceTypePermission:
		permission, ok := resource.(models.Permission)
		if !ok {
			return invalidResourceValueTypeError
		}
		return r.permissionsCollection.DeletePermission(tenantID, permission.ID.String())
	default:
		return erp_errors.Validation(erp_errors.ValidationInvalidValue).WithError(fmt.Errorf("unsupported resource type"))
	}
}

func (r *RBACManager) GetResource(tenantID string, userID string, resourceType string, resourceID string) (any, error) {
	permission := fmt.Sprintf("%s:%s", resourceType, models.PermissionActionRead)
	if err := r.hasPermission(tenantID, userID, permission); err != nil {
		return nil, err
	}
	switch resourceType {
	case models.ResourceTypeUser:
		return r.userCollection.GetUserByID(tenantID, resourceID)
	case models.ResourceTypeRole:
		return r.rolesCollection.GetRoleByID(tenantID, resourceID)
	case models.ResourceTypePermission:
		return r.permissionsCollection.GetPermissionByID(tenantID, resourceID)
	default:
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue).WithError(fmt.Errorf("unsupported resource type"))
	}
}

func (r *RBACManager) GetResources(tenantID string, userID string, resourceType string) ([]any, error) {
	permission := fmt.Sprintf("%s:%s", resourceType, models.PermissionActionRead)
	if err := r.hasPermission(tenantID, userID, permission); err != nil {
		return nil, err
	}
	switch resourceType {
	case models.ResourceTypeUser:
		users, err := r.userCollection.GetUsersByTenantID(tenantID)
		if err != nil {
			return nil, err
		}
		usersAny := make([]any, 0)
		for _, user := range users {
			usersAny = append(usersAny, user)
		}
		return usersAny, nil
	case models.ResourceTypeRole:
		roles, err := r.rolesCollection.GetRolesByTenantID(tenantID)
		if err != nil {
			return nil, err
		}
		rolesAny := make([]any, 0)
		for _, role := range roles {
			rolesAny = append(rolesAny, role)
		}
		return rolesAny, nil
	case models.ResourceTypePermission:
		permissions, err := r.permissionsCollection.GetPermissionsByTenantID(tenantID)
		if err != nil {
			return nil, err
		}
		permissionsAny := make([]any, 0)
		for _, permission := range permissions {
			permissionsAny = append(permissionsAny, permission)
		}
		return permissionsAny, nil
	default:
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue).WithError(fmt.Errorf("unsupported resource type"))
	}
}

/* Get Functions */
func (r *RBACManager) GetUserPermissions(tenantID string, userID string) (map[string]bool, error) {
	user, err := r.userCollection.GetUserByID(tenantID, userID)
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
	return userPermissions, nil
}

func (r *RBACManager) GetUserRoles(tenantID string, userID string) ([]string, error) {
	user, err := r.userCollection.GetUserByID(tenantID, userID)
	if err != nil {
		return nil, err
	}
	userRoles := make([]string, 0)
	for _, userRole := range user.Roles {
		userRoles = append(userRoles, userRole.RoleID)
	}
	return userRoles, nil
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
	// Verify permissions format
	for _, permission := range permissions {
		if !models.IsValidPermissionFormat(permission) {
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

/* Helper Functions */
func (r *RBACManager) hasPermission(tenantID string, userID string, permission string) error {
	permissionsCheckResponse, err := r.CheckUserPermissions(tenantID, userID, []string{permission})
	if err != nil {
		return err
	}
	if !permissionsCheckResponse[permission] {
		return erp_errors.Auth(erp_errors.AuthPermissionDenied)
	}
	return nil
}
