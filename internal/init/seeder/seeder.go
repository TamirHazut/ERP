package seeder

import (
	"fmt"

	collection_auth "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/auth/hash"
	"erp.localhost/internal/infra/db"
	mongo_db "erp.localhost/internal/infra/db/mongo"
	"erp.localhost/internal/infra/db/redis"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Seeder struct {
	logger logger.Logger

	// Handlers for database operations
	tenantHandler    *collection_auth.TenantCollection
	userHandler      *collection_auth.UserCollection
	roleHandler      *collection_auth.RoleCollection
	systemKeyHandler *redis.SystemKeyHandler
}

func NewSeeder(logger logger.Logger) (*Seeder, *infra_error.AppError) {
	th, err := collection_auth.NewTenantCollection(logger)
	if err != nil {
		logger.Fatal("failed to create tenant collection", "error", err)
		return nil, err
	}
	uh, err := collection_auth.NewUserCollection(logger)
	if err != nil {
		logger.Fatal("failed to create user collection", "error", err)
		return nil, err
	}
	rh, err := collection_auth.NewRoleCollection(logger)
	if err != nil {
		logger.Fatal("failed to create role collection", "error", err)
		return nil, err
	}
	systemKeyHandler, err := redis.NewSystemKeyHandler(logger)
	if err != nil {
		logger.Fatal("failed to system key handler", "error", err)
		return nil, err
	}
	return &Seeder{
		logger:           logger,
		tenantHandler:    th,
		userHandler:      uh,
		roleHandler:      rh,
		systemKeyHandler: systemKeyHandler,
	}, nil
}

func (s *Seeder) SeedSystemData() *infra_error.AppError {
	s.logger.Info("Seeding system data")

	// Step 0: Create indexes BEFORE seeding data
	if err := s.SeedIndexes(); err != nil {
		return err
	}

	// Step 1: Create system tenant
	systemTenantID, err := s.seedSystemTenant()
	if err != nil {
		return err
	}
	s.logger.Info("System tenant seeded")

	// Step 2: Create system role
	roleID, err := s.seedSystemRole(systemTenantID, db.TenantAdminPermission)
	if err != nil {
		return err
	}
	s.logger.Info("System role seeded")

	// Step 4: Create system admin user
	_, err = s.seedSystemAdminUser(systemTenantID, roleID)
	if err != nil {
		return err
	}
	s.logger.Info("System admin user seeded")

	return nil
}

// SeedIndexes ensures all indexes are created for system collections
func (s *Seeder) SeedIndexes() *infra_error.AppError {
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
		// {
		// 	dbName:     model_mongo.EventDB,
		// 	collection: model_mongo.AuditLogsCollection,
		// 	indexes:    model_mongo.GetAuditLogsIndexes(),
		// },
	}

	// Create indexes for each collection
	for _, mapping := range indexMappings {
		dbManager, err := mongo_db.NewMongoDBManager(mapping.dbName, s.logger)
		if err != nil {
			s.logger.Error(fmt.Sprintf("failed to create DB manager for %s", mapping.dbName), "error", err)
			return err
		}
		defer dbManager.Close()

		if err := dbManager.EnsureIndexes(string(mapping.collection), mapping.indexes); err != nil {
			s.logger.Error(fmt.Sprintf("failed to create indexes for %s.%s", mapping.dbName, mapping.collection), "error", err)
			return err
		}
	}

	s.logger.Info("All indexes created successfully")
	return nil
}

func (s *Seeder) seedSystemTenant() (string, *infra_error.AppError) {
	s.logger.Debug("Checking for existing system tenant")
	// Check if system tenant already exists
	filter := map[string]any{"name": db.SystemTenant}
	existing, err := s.tenantHandler.FindOne(filter)
	if err == nil && existing != nil {
		s.logger.Info("System tenant already exists, skipping creation")
		return existing.Id, s.systemKeyHandler.Set("tenant", &existing.Id)
	}

	s.logger.Debug("Creating system tenant")

	// Create new tenant
	tenant := &authv1.Tenant{
		Name:      db.SystemTenant,
		Status:    authv1.TenantStatus_TENANT_STATUS_ACTIVE,
		CreatedBy: "System",
		Protected: true,
	}

	tenantID, err := s.tenantHandler.Create(tenant)
	if err != nil {
		return "", infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	return tenantID, s.systemKeyHandler.Set("tenant", &tenantID)
}

func (s *Seeder) seedSystemRole(systemTenantID, permissionString string) (string, *infra_error.AppError) {
	s.logger.Debug("Checking for existing system role")

	// Check if role already exists
	filter := map[string]any{
		"tenant_id": systemTenantID,
		"name":      db.SystemAdminUser,
	}
	existing, err := s.roleHandler.FindOne(filter)
	if err == nil && existing != nil {
		s.logger.Info("System role already exists, skipping creation")
		return existing.Id, s.systemKeyHandler.Set("role", &existing.Id)
	}

	s.logger.Debug("Creating system role")

	// Create role
	role := &authv1.Role{
		TenantId:    systemTenantID,
		Name:        db.SystemAdminUser,
		Description: "System administrator role with full access to all resources",
		Permissions: []string{permissionString},
		Status:      authv1.RoleStatus_ROLE_STATUS_ACTIVE,
		CreatedBy:   "System",
		Protected:   true,
	}

	roleID, err := s.roleHandler.Create(role)
	if err != nil {
		return "", infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	return roleID, s.systemKeyHandler.Set("role", &roleID)
}

func (s *Seeder) seedSystemAdminUser(systemTenantID, systemRoleID string) (string, *infra_error.AppError) {
	s.logger.Debug("Checking for existing system admin user")

	// Check if user already exists
	filter := map[string]any{
		"tenant_id": systemTenantID,
		"email":     db.SystemAdminEmail,
	}
	existing, err := s.userHandler.FindOne(filter)
	if err == nil && existing != nil {
		s.logger.Info("System admin user already exists, skipping creation")
		return existing.Id, s.systemKeyHandler.Set("user", &existing.Id)
	}

	s.logger.Debug("Creating system admin user")

	// Hash password
	hash, err := hash.Hash(db.SystemAdminPassword)
	if err != nil {
		return "", infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}

	// Create user with system admin role
	user := &authv1.User{
		TenantId:  systemTenantID,
		Username:  db.SystemAdminUser,
		Email:     db.SystemAdminEmail,
		Password:  hash,
		Status:    authv1.UserStatus_USER_STATUS_ACTIVE,
		CreatedBy: "System",
		Protected: true,
		Roles: []*authv1.UserRole{
			{
				TenantId:   systemTenantID,
				RoleId:     systemRoleID,
				AssignedAt: timestamppb.Now(),
				AssignedBy: "System",
			},
		},
	}

	userID, err := s.userHandler.Create(user)
	if err != nil {
		return "", infra_error.Internal(infra_error.InternalDatabaseError, err)
	}

	return userID, s.systemKeyHandler.Set("user", &userID)
}
