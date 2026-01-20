package seeder

import (
	"fmt"

	"erp.localhost/internal/auth/hash"
	"erp.localhost/internal/infra/db"
	mongo_db "erp.localhost/internal/infra/db/mongo"
	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	"erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Seeder struct {
	logger logger.Logger

	// Handlers for database operations
	tenantHandler     *collection.BaseCollectionHandler[authv1.Tenant]
	userHandler       *collection.BaseCollectionHandler[authv1.User]
	permissionHandler *collection.BaseCollectionHandler[authv1.Permission]
	roleHandler       *collection.BaseCollectionHandler[authv1.Role]
}

func NewSeeder(logger logger.Logger) *Seeder {
	return &Seeder{
		logger: logger,
		tenantHandler: collection.NewBaseCollectionHandler[authv1.Tenant](
			model_mongo.AuthDB,
			model_mongo.TenantsCollection,
			logger,
		),
		userHandler: collection.NewBaseCollectionHandler[authv1.User](
			model_mongo.AuthDB,
			model_mongo.UsersCollection,
			logger,
		),
		permissionHandler: collection.NewBaseCollectionHandler[authv1.Permission](
			model_mongo.AuthDB,
			model_mongo.PermissionsCollection,
			logger,
		),
		roleHandler: collection.NewBaseCollectionHandler[authv1.Role](
			model_mongo.AuthDB,
			model_mongo.RolesCollection,
			logger,
		),
	}
}

func (s *Seeder) SeedSystemData() error {
	s.logger.Info("Seeding system data")

	// Step 0: Create indexes BEFORE seeding data
	if err := s.SeedIndexes(); err != nil {
		return fmt.Errorf("failed to seed indexes: %w", err)
	}

	// Step 1: Create system tenant
	if err := s.seedSystemTenant(); err != nil {
		return fmt.Errorf("failed to seed system tenant: %w", err)
	}
	s.logger.Info("System tenant seeded", "tenant_id", db.SystemTenantID)

	// Step 2: Create system permission
	if err := s.seedSystemPermission(); err != nil {
		return fmt.Errorf("failed to seed system permission: %w", err)
	}
	s.logger.Info("System permission seeded", "permission_id", db.SystemAdminPermissionID)

	// Step 3: Create system role
	if err := s.seedSystemRole(); err != nil {
		return fmt.Errorf("failed to seed system role: %w", err)
	}
	s.logger.Info("System role seeded", "role_id", db.SystemAdminRoleID)

	// Step 4: Create system admin user
	if err := s.seedSystemAdminUser(); err != nil {
		return fmt.Errorf("failed to seed system admin user: %w", err)
	}
	s.logger.Info("System admin user seeded", "user_id", db.SystemAdminUserID)

	return nil
}

// SeedIndexes ensures all indexes are created for system collections
func (s *Seeder) SeedIndexes() error {
	s.logger.Info("Creating indexes for system collections")

	// Define collection-index mappings
	indexMappings := []struct {
		dbName     model_mongo.DBName
		collection model_mongo.Collection
		indexes    []mongo.IndexModel
	}{
		{
			dbName:     model_mongo.AuthDB,
			collection: model_mongo.TenantsCollection,
			indexes:    model_mongo.GetTenantsIndexes(),
		},
		{
			dbName:     model_mongo.AuthDB,
			collection: model_mongo.UsersCollection,
			indexes:    model_mongo.GetUsersIndexes(),
		},
		{
			dbName:     model_mongo.AuthDB,
			collection: model_mongo.RolesCollection,
			indexes:    model_mongo.GetRolesIndexes(),
		},
		{
			dbName:     model_mongo.AuthDB,
			collection: model_mongo.PermissionsCollection,
			indexes:    model_mongo.GetPermissionsIndexes(),
		},
		{
			dbName:     model_mongo.EventDB,
			collection: model_mongo.AuditLogsCollection,
			indexes:    model_mongo.GetAuditLogsIndexes(),
		},
	}

	// Create indexes for each collection
	for _, mapping := range indexMappings {
		dbManager := mongo_db.NewMongoDBManager(mapping.dbName, s.logger)
		if dbManager == nil {
			return fmt.Errorf("failed to create DB manager for %s", mapping.dbName)
		}
		defer dbManager.Close()

		if err := dbManager.EnsureIndexes(string(mapping.collection), mapping.indexes); err != nil {
			return fmt.Errorf("failed to create indexes for %s.%s: %w",
				mapping.dbName, mapping.collection, err)
		}
	}

	s.logger.Info("All indexes created successfully")
	return nil
}

func (s *Seeder) seedSystemTenant() error {
	s.logger.Debug("Checking for existing system tenant")
	// Check if system tenant already exists
	filter := map[string]any{"name": db.SystemTenant}
	existing, err := s.tenantHandler.FindOne(filter)
	if err == nil && existing != nil {
		s.logger.Info("System tenant already exists, skipping creation")
		db.SystemTenantID = existing.Id
		return nil
	}

	s.logger.Debug("Creating system tenant")

	// Create new tenant
	tenant := &authv1.Tenant{
		Name:      db.SystemTenant,
		Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
		CreatedBy: "System",
	}

	tenantID, err := s.tenantHandler.Create(tenant)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	db.SystemTenantID = tenantID
	return nil
}

func (s *Seeder) seedSystemPermission() error {
	s.logger.Debug("Checking for existing system permission")

	permissionString := fmt.Sprintf("%s:%s", auth.ResourceTypeAll, auth.PermissionActionAll)

	// Check if permission already exists
	filter := map[string]any{
		"tenant_id":         db.SystemTenantID,
		"permission_string": permissionString,
	}
	existing, err := s.permissionHandler.FindOne(filter)
	if err == nil && existing != nil {
		s.logger.Info("System permission already exists, skipping creation")
		db.SystemAdminPermissionID = existing.Id
		return nil
	}

	s.logger.Debug("Creating system permission")

	// Create permission
	permission := &authv1.Permission{
		TenantId:         db.SystemTenantID,
		Resource:         auth.ResourceTypeAll,
		Action:           auth.PermissionActionAll,
		CreatedBy:        "System",
		DisplayName:      "System Controller",
		Description:      "Full system access - all resources and actions",
		PermissionString: permissionString,
		IsDangerous:      true,
	}

	permissionID, err := s.permissionHandler.Create(permission)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	db.SystemAdminPermissionID = permissionID
	return nil
}

func (s *Seeder) seedSystemRole() error {
	s.logger.Debug("Checking for existing system role")

	// Check if role already exists
	filter := map[string]any{
		"tenant_id": db.SystemTenantID,
		"name":      db.SystemAdminUser,
	}
	existing, err := s.roleHandler.FindOne(filter)
	if err == nil && existing != nil {
		s.logger.Info("System role already exists, skipping creation")
		db.SystemAdminRoleID = existing.Id
		return nil
	}

	s.logger.Debug("Creating system role")

	// Create role
	role := &authv1.Role{
		TenantId:    db.SystemTenantID,
		Name:        db.SystemAdminUser,
		Description: "System administrator role with full access to all resources",
		Permissions: []string{db.SystemAdminPermissionID},
		Status:      authv1.RoleStatus_ROLE_STATUS_ACTIVE,
		CreatedBy:   "System",
	}

	roleID, err := s.roleHandler.Create(role)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	db.SystemAdminRoleID = roleID
	return nil
}

func (s *Seeder) seedSystemAdminUser() error {
	s.logger.Debug("Checking for existing system admin user")

	// Check if user already exists
	filter := map[string]any{
		"tenant_id": db.SystemTenantID,
		"email":     db.SystemAdminEmail,
	}
	existing, err := s.userHandler.FindOne(filter)
	if err == nil && existing != nil {
		s.logger.Info("System admin user already exists, skipping creation")
		db.SystemAdminUserID = existing.Id
		return nil
	}

	s.logger.Debug("Creating system admin user")

	// Hash password
	hash, err := hash.HashPassword(db.SystemAdminPassword)
	if err != nil {
		return infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}

	// Create user with system admin role
	user := &authv1.User{
		TenantId:     db.SystemTenantID,
		Username:     db.SystemAdminUser,
		Email:        db.SystemAdminEmail,
		PasswordHash: hash,
		Status:       authv1.UserStatus_USER_STATUS_ACTIVE,
		CreatedBy:    "System",
		Roles: []*authv1.UserRole{
			{
				TenantId:   db.SystemTenantID,
				RoleId:     db.SystemAdminRoleID,
				AssignedAt: timestamppb.Now(),
				AssignedBy: "System",
			},
		},
	}

	userID, err := s.userHandler.Create(user)
	if err != nil {
		return infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	db.SystemAdminUserID = userID
	return nil
}
