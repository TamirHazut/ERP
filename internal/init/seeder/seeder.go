package seeder

import (
	"fmt"
	"time"

	"erp.localhost/internal/auth/password"
	"erp.localhost/internal/infra/db"
	mongo_db "erp.localhost/internal/infra/db/mongo"
	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

type Seeder struct {
	logger logger.Logger

	// Handlers for database operations
	tenantHandler     *collection.BaseCollectionHandler[model_auth.Tenant]
	userHandler       *collection.BaseCollectionHandler[model_auth.User]
	permissionHandler *collection.BaseCollectionHandler[model_auth.Permission]
	roleHandler       *collection.BaseCollectionHandler[model_auth.Role]
}

func NewSeeder(logger logger.Logger) *Seeder {
	return &Seeder{
		logger: logger,
		tenantHandler: collection.NewBaseCollectionHandler[model_auth.Tenant](
			model_mongo.AuthDB,
			model_mongo.TenantsCollection,
			logger,
		),
		userHandler: collection.NewBaseCollectionHandler[model_auth.User](
			model_mongo.AuthDB,
			model_mongo.UsersCollection,
			logger,
		),
		permissionHandler: collection.NewBaseCollectionHandler[model_auth.Permission](
			model_mongo.AuthDB,
			model_mongo.PermissionsCollection,
			logger,
		),
		roleHandler: collection.NewBaseCollectionHandler[model_auth.Role](
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
			dbName:     model_mongo.AuthDB,
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
		db.SystemTenantID = existing.ID.Hex()
		return nil
	}

	s.logger.Debug("Creating system tenant")

	// Create new tenant
	tenant := &model_auth.Tenant{
		Name:      db.SystemTenant,
		Status:    model_auth.TenantStatusActive,
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

	permissionString := fmt.Sprintf("%s:%s", model_auth.ResourceTypeAll, model_auth.PermissionActionAll)

	// Check if permission already exists
	filter := map[string]any{
		"tenant_id":         db.SystemTenantID,
		"permission_string": permissionString,
	}
	existing, err := s.permissionHandler.FindOne(filter)
	if err == nil && existing != nil {
		s.logger.Info("System permission already exists, skipping creation")
		db.SystemAdminPermissionID = existing.ID.Hex()
		return nil
	}

	s.logger.Debug("Creating system permission")

	// Create permission
	permission := &model_auth.Permission{
		TenantID:         db.SystemTenantID,
		Resource:         model_auth.ResourceTypeAll,
		Action:           model_auth.PermissionActionAll,
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
		db.SystemAdminRoleID = existing.ID.Hex()
		return nil
	}

	s.logger.Debug("Creating system role")

	// Create role
	role := &model_auth.Role{
		TenantID:      db.SystemTenantID,
		Name:          db.SystemAdminUser,
		Description:   "System administrator role with full access to all resources",
		IsTenantAdmin: true,
		Permissions:   []string{db.SystemAdminPermissionID},
		Status:        model_auth.RoleStatusActive,
		CreatedBy:     "System",
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
		db.SystemAdminUserID = existing.ID.Hex()
		return nil
	}

	s.logger.Debug("Creating system admin user")

	// Hash password
	hash, err := password.HashPassword(db.SystemAdminPassword)
	if err != nil {
		return infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}

	// Create user with system admin role
	user := &model_auth.User{
		TenantID:     db.SystemTenantID,
		Username:     db.SystemAdminUser,
		Email:        db.SystemAdminEmail,
		PasswordHash: hash,
		Status:       model_auth.UserStatusActive,
		CreatedBy:    "System",
		Roles: []model_auth.UserRole{
			{
				TenantID:   db.SystemTenantID,
				RoleID:     db.SystemAdminRoleID,
				AssignedAt: time.Now(),
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
