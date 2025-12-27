package rbac

import (
	collection "erp.localhost/internal/auth/collections"
	"erp.localhost/internal/db/mongo"
	"erp.localhost/internal/logging"
)

type RBACManager struct {
	logger                *logging.Logger
	userCollection        *collection.UserCollection
	rolesCollection       *collection.RolesCollection
	permissionsCollection *collection.PermissionsCollection
}

func NewRBACManager() *RBACManager {
	logger := logging.NewLogger(logging.ModuleAuth)
	dbHandler := mongo.NewMongoDBManager(mongo.AuthDB)
	if dbHandler == nil {
		logger.Fatal("failed to create db handler")
		return nil
	}
	return &RBACManager{
		logger:                logger,
		userCollection:        collection.NewUserCollection(dbHandler),
		rolesCollection:       collection.NewRoleCollection(dbHandler),
		permissionsCollection: collection.NewPermissionCollection(dbHandler),
	}
}

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

func (r *RBACManager) CheckUserPermissions(tenantID string, userID string, permissions []string) (map[string]bool, error) {
	userPermissions, err := r.GetUserPermissions(tenantID, userID)
	if err != nil {
		return nil, err
	}
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
