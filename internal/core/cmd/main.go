package cmd

import (
	"fmt"
	"time"

	"erp.localhost/internal/auth/models"
	auth "erp.localhost/internal/auth/utils"
	common_models "erp.localhost/internal/common/models"
	"erp.localhost/internal/db"
	"erp.localhost/internal/db/mongo"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

var (
	permissionAllString = fmt.Sprintf("%s:%s", models.ResourceTypeAll, models.PermissionActionAll)
)

func Main() {
	if err := createDefaultData(); err != nil {
		return
	}
}

func createDefaultData() error {
	logger := logging.NewLogger(common_models.ModuleAuth)

	logger.Debug("Creating default data")

	logger.Debug("Creating system tenant")
	if err := createSystemTenant(mongo.NewBaseCollectionHandler[models.Tenant](string(mongo.TenantsCollection), logging.NewLogger(common_models.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system tenant", "error", err)
		return err
	}
	logger.Debug("System tenant created")

	logger.Debug("Creating system admin role")
	if err := createSystemAdminRole(mongo.NewBaseCollectionHandler[models.Role](string(mongo.RolesCollection), logging.NewLogger(common_models.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system admin role", "error", err)
		return err
	}
	logger.Debug("System admin role created")

	logger.Debug("Creating system admin permission")
	if err := createSystemAdminPermission(mongo.NewBaseCollectionHandler[models.Permission](string(mongo.PermissionsCollection), logging.NewLogger(common_models.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system admin permission", "error", err)
		return err
	}
	logger.Debug("System admin permission created")

	logger.Debug("Creating system admin user")
	if err := createSystemAdminUser(mongo.NewBaseCollectionHandler[models.User](string(mongo.UsersCollection), logging.NewLogger(common_models.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system admin user", "error", err)
		return err
	}
	logger.Debug("System admin user created")
	return nil
}

func createSystemTenant(collection mongo.CollectionHandler[models.Tenant]) error {
	tenant := models.Tenant{
		Name:      "System",
		Status:    models.TenantStatusActive,
		CreatedBy: "System",
	}
	systemTenantID, err := collection.Create(tenant)
	if err != nil || systemTenantID == "" {
		return erp_errors.Internal(erp_errors.InternalUnexpectedError, err)
	}
	db.SystemTenantID = systemTenantID
	return nil
}

func createSystemAdminRole(collection mongo.CollectionHandler[models.Role]) error {
	role := models.Role{
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

func createSystemAdminPermission(collection mongo.CollectionHandler[models.Permission]) error {

	permission := models.Permission{
		TenantID:         db.SystemTenantID,
		Resource:         models.ResourceTypeAll,
		Action:           models.PermissionActionAll,
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

func createSystemAdminUser(collectionHandler mongo.CollectionHandler[models.User]) error {
	hash, _ := auth.HashPassword(db.SystemAdminPassword)
	user := models.User{
		TenantID:     db.SystemTenantID,
		Username:     db.SystemAdminUser,
		Email:        db.SystemAdminEmail,
		PasswordHash: hash,
		Status:       models.UserStatusActive,
		CreatedBy:    "System",
		Roles: []models.UserRole{
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
