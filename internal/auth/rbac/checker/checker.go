package checker

import (
	"fmt"
	"slices"

	"erp.localhost/internal/auth/collection"
	infra_error "erp.localhost/internal/infra/error"
	model_auth "erp.localhost/internal/infra/model/auth"
)

//go:generate mockgen -destination=mock/mock_rbac_checker.go -package=mock erp.localhost/internal/auth/rbac/checker RBACChecker
type RBACChecker interface {
	CheckUserPermissions(permissions []string, userPermissions map[string]bool) (map[string]bool, error)
	CheckUserRoles(roles []string, userRoles []string) (map[string]bool, error)
	VerifyPerimissionsFormat(permissions []string) error
}

type BaseRBACChecker struct {
	permissionsCollection *collection.PermissionsCollection
	rolesCollection       *collection.RolesCollection
}

func NewBaseRBACChecker(permissionsCollection *collection.PermissionsCollection, rolesCollection *collection.RolesCollection) *BaseRBACChecker {
	return &BaseRBACChecker{
		permissionsCollection: permissionsCollection,
		rolesCollection:       rolesCollection,
	}
}

func (r *BaseRBACChecker) CheckUserPermissions(permissions []string, userPermissions map[string]bool) (map[string]bool, error) {
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

func (r *BaseRBACChecker) CheckUserRoles(roles []string, userRoles []string) (map[string]bool, error) {
	// Check requested permissions status
	rolesCheckResponse := make(map[string]bool, 0)
	for _, role := range roles {
		rolesCheckResponse[role] = slices.Contains(userRoles, role)
	}
	return rolesCheckResponse, nil
}
func (r *BaseRBACChecker) VerifyUserRole(tenantID string, userID string, roleID string) (bool, error) {
	return false, nil
}
func (r *BaseRBACChecker) VerifyRolePermissions(tenantID string, roleID string, permissions []string) (map[string]bool, error) {
	return nil, nil
}

func (r *BaseRBACChecker) VerifyPerimissionsFormat(permissions []string) error {
	// Verify permissions format
	for _, permission := range permissions {
		if !model_auth.IsValidPermissionFormat(permission) {
			return infra_error.Validation(infra_error.ValidationInvalidValue).WithError(fmt.Errorf("invalid permission format %s", permission))
		}
	}
	return nil
}

func (r *BaseRBACChecker) HasPermission(tenantID string, userID string, permission string) error {
	return nil
}
