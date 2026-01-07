package rbac

import (
	"errors"
	"fmt"

	collection "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/auth/rbac/checker"
	"erp.localhost/internal/auth/rbac/handler"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_shared "erp.localhost/internal/infra/model/shared"
)

type RBACManager struct {
	logger       logger.Logger
	rbacHandlers map[string]handler.RBACResourceHandler
	rbacChecker  checker.RBACChecker
	// userGRPCClient proto_core.UserServiceClient
}

func NewRBACManager(permissionsCollection *collection.PermissionsCollection, rolesCollection *collection.RolesCollection) *RBACManager {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)

	return &RBACManager{
		logger: logger,
		rbacHandlers: map[string]handler.RBACResourceHandler{
			model_auth.ResourceTypePermission: handler.NewPermissionHandler(permissionsCollection),
			model_auth.ResourceTypeRole:       handler.NewRoleHandler(rolesCollection),
		},
		rbacChecker: checker.NewBaseRBACChecker(permissionsCollection, rolesCollection),
	}
}

/* CRUD Resource Functions */
func (r *RBACManager) CreateResource(tenantID string, userID string, resourceType string, resource model_auth.RBACResource) (string, error) {
	permission := fmt.Sprintf("%s:%s", resourceType, model_auth.PermissionActionCreate)

	if err := r.HasPermission(tenantID, userID, permission); err != nil {
		return "", err
	}
	resourceHandler, ok := r.rbacHandlers[resourceType]
	if !ok {
		return "", infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType")
	}
	return resourceHandler.CreateResource(resource)
}

func (r *RBACManager) UpdateResource(tenantID string, userID string, resourceType string, resource model_auth.RBACResource) error {
	permission := fmt.Sprintf("%s:%s", resourceType, model_auth.PermissionActionUpdate)

	if err := r.HasPermission(tenantID, userID, permission); err != nil {
		return err
	}

	resourceHandler, ok := r.rbacHandlers[resourceType]
	if !ok {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType")
	}
	return resourceHandler.UpdateResource(resource)
}

func (r *RBACManager) DeleteResource(tenantID string, userID string, resourceType string, resourceID string) error {
	permission := fmt.Sprintf("%s:%s", resourceType, model_auth.PermissionActionDelete)

	if err := r.HasPermission(tenantID, userID, permission); err != nil {
		return err
	}

	resourceHandler, ok := r.rbacHandlers[resourceType]
	if !ok {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType")
	}
	return resourceHandler.DeleteResource(tenantID, resourceID)
}

func (r *RBACManager) GetResource(tenantID string, userID string, resourceType string, resourceID string) (model_auth.RBACResource, error) {
	permission := fmt.Sprintf("%s:%s", resourceType, model_auth.PermissionActionRead)
	if err := r.HasPermission(tenantID, userID, permission); err != nil {
		return nil, err
	}
	resourceHandler, ok := r.rbacHandlers[resourceType]
	if !ok {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType")
	}

	return resourceHandler.GetResource(tenantID, handler.SearchFilter{handler.FilterKeyResourceID: resourceID})
}

// TODO: add support in pagination and filter
func (r *RBACManager) GetResources(tenantID string, userID string, resourceType string) ([]model_auth.RBACResource, error) {
	permission := fmt.Sprintf("%s:%s", resourceType, model_auth.PermissionActionRead)
	if err := r.HasPermission(tenantID, userID, permission); err != nil {
		return nil, err
	}
	resourceHandler, ok := r.rbacHandlers[resourceType]
	if !ok {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType")
	}
	return resourceHandler.ListResources(tenantID, nil)
}

/* Get Functions */
// TODO: uncomment when user grpc service is ready
func (r *RBACManager) getUserPermissions(tenantID string, userID string) (map[string]bool, error) {
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
	return nil, infra_error.Internal(infra_error.InternalServiceUnavailable, errors.New("unimplemented function"))
}

// TODO: uncomment when user grpc service is ready
func (r *RBACManager) getUserRoles(tenantID string, userID string) ([]string, error) {
	// reqOpts := &proto_core.GetUserRequest{
	// 	Identifier: &infra_proto.UserIdentifier{
	// 		TenantId: tenantID,
	// 		UserId: requestorUserID,
	// 	},
	// }
	// user, err := r.userGRPCClient.GetUser(context.Background(), reqOpts)
	// if err != nil {
	// 	return nil, err
	// }
	// convertor

	// return roles
	// userRoles := make([]string, 0)
	// for _, userRole := range user.Roles {
	// 	userRoles = append(userRoles, userRole.RoleID)
	// }
	// return userRoles, nil
	return nil, infra_error.Internal(infra_error.InternalServiceUnavailable, errors.New("unimplemented function"))
}

// /* Verification Functions */
func (r *RBACManager) CheckUserPermissions(tenantID string, userID string, permissions []string) (map[string]bool, error) {
	if tenantID == "" || userID == "" || len(permissions) == 0 {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenantID", "userID", "permission")
	}
	if err := r.rbacChecker.VerifyPerimissionsFormat(permissions); err != nil {
		return nil, err
	}
	// Get user permissions
	userPermissions, err := r.getUserPermissions(tenantID, userID)
	if err != nil {
		return nil, err
	}
	return r.rbacChecker.CheckUserPermissions(permissions, userPermissions)
}

func (r *RBACManager) CheckUserRoles(tenantID string, userID string, roles []string) (map[string]bool, error) {
	if tenantID == "" || userID == "" || len(roles) == 0 {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenantID", "userID", "roles")
	}

	// Get user permissions
	userRoles, err := r.getUserRoles(tenantID, userID)
	if err != nil {
		return nil, err
	}
	return r.rbacChecker.CheckUserRoles(roles, userRoles)
}

func (r *RBACManager) HasPermission(tenantID string, userID string, permission string) error {
	permissionsCheckResponse, err := r.CheckUserPermissions(tenantID, userID, []string{permission})
	if err != nil {
		return err
	}
	if !permissionsCheckResponse[permission] {
		return infra_error.Auth(infra_error.AuthPermissionDenied)
	}
	return nil
}
