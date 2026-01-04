package cmd

import (
	"fmt"
	"time"

	auth "erp.localhost/internal/auth/utils"
	"erp.localhost/internal/infra/db"
	"erp.localhost/internal/infra/db/mongo"
	erp_errors "erp.localhost/internal/infra/errors"
	"erp.localhost/internal/infra/logging"
	auth_models "erp.localhost/internal/infra/models/auth"
	core_models "erp.localhost/internal/infra/models/core"
	mongo_models "erp.localhost/internal/infra/models/db/mongo"
	shared_models "erp.localhost/internal/infra/models/shared"
)

var (
	permissionAllString = fmt.Sprintf("%s:%s", auth_models.ResourceTypeAll, auth_models.PermissionActionAll)
)

func Main() {
	if err := createDefaultData(); err != nil {
		return
	}
}

func createDefaultData() error {
	logger := logging.NewLogger(shared_models.ModuleAuth)

	logger.Debug("Creating default data")

	logger.Debug("Creating system tenant")
	if err := createSystemTenant(mongo.NewBaseCollectionHandler[core_models.Tenant](string(mongo_models.TenantsCollection), logging.NewLogger(shared_models.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system tenant", "error", err)
		return err
	}
	logger.Debug("System tenant created")

	logger.Debug("Creating system admin role")
	if err := createSystemAdminRole(mongo.NewBaseCollectionHandler[auth_models.Role](string(mongo_models.RolesCollection), logging.NewLogger(shared_models.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system admin role", "error", err)
		return err
	}
	logger.Debug("System admin role created")

	logger.Debug("Creating system admin permission")
	if err := createSystemAdminPermission(mongo.NewBaseCollectionHandler[auth_models.Permission](string(mongo_models.PermissionsCollection), logging.NewLogger(shared_models.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system admin permission", "error", err)
		return err
	}
	logger.Debug("System admin permission created")

	logger.Debug("Creating system admin user")
	if err := createSystemAdminUser(mongo.NewBaseCollectionHandler[core_models.User](string(mongo_models.UsersCollection), logging.NewLogger(shared_models.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system admin user", "error", err)
		return err
	}
	logger.Debug("System admin user created")
	return nil
}

func createSystemTenant(collection mongo.CollectionHandler[core_models.Tenant]) error {
	tenant := core_models.Tenant{
		Name:      "System",
		Status:    auth_models.TenantStatusActive,
		CreatedBy: "System",
	}
	systemTenantID, err := collection.Create(tenant)
	if err != nil || systemTenantID == "" {
		return erp_errors.Internal(erp_errors.InternalUnexpectedError, err)
	}
	db.SystemTenantID = systemTenantID
	return nil
}

func createSystemAdminRole(collection mongo.CollectionHandler[auth_models.Role]) error {
	role := auth_models.Role{
		Name:        "SystemAdmin",
		Description: "System admin role",
		Permissions: []string{permissionAllString},
	}
	roleID, err := collection.Create(role)
	if err != nil || roleID == "" {
		return erp_errors.Internal(erp_errors.InternalUnexpectedError, err)
	}
	db.SystemAdminRoleID = roleID
	return nil
}

func createSystemAdminPermission(collection mongo.CollectionHandler[auth_models.Permission]) error {

	permission := auth_models.Permission{
		TenantID:         db.SystemTenantID,
		Resource:         auth_models.ResourceTypeAll,
		Action:           auth_models.PermissionActionAll,
		CreatedBy:        "System",
		DisplayName:      "System Controller",
		PermissionString: permissionAllString,
		IsDangerous:      true,
	}
	permissionID, err := collection.Create(permission)
	if err != nil || permissionID == "" {
		return erp_errors.Internal(erp_errors.InternalUnexpectedError, err)
	}
	db.SystemAdminPermissionID = permissionID
	return nil
}

func createSystemAdminUser(collectionHandler mongo.CollectionHandler[core_models.User]) error {
	hash, _ := auth.HashPassword(db.SystemAdminPassword)
	user := core_models.User{
		TenantID:     db.SystemTenantID,
		Username:     db.SystemAdminUser,
		Email:        db.SystemAdminEmail,
		PasswordHash: hash,
		Status:       auth_models.UserStatusActive,
		CreatedBy:    "System",
		Roles: []core_models.UserRole{
			{
				TenantID:   db.SystemTenantID,
				RoleID:     db.SystemAdminRoleID,
				AssignedAt: time.Now(),
				AssignedBy: "System",
			},
		},
	}
	userID, err := collectionHandler.Create(user)
	if err != nil || userID == "" {
		return erp_errors.Internal(erp_errors.InternalUnexpectedError, err)
	}
	db.SystemAdminUserID = userID
	return nil
}
