package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"erp.localhost/internal/core/service"
	grpc_server "erp.localhost/internal/infra/grpc/server"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_core "erp.localhost/internal/infra/proto/core/v1"
	"google.golang.org/grpc"
)

const (
	serverPort = 5001
)

func Main() {
	// Channel to listen for OS signals for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to signal the gRPC server goroutine to stop
	quit := make(chan struct{})

	// Create gRPC Server
	certs := model_shared.NewCerts()
	if certs == nil {
		return
	}
	services := map[*grpc.ServiceDesc]any{
		&proto_core.UserService_ServiceDesc:   service.NewUserService(),
		&proto_core.TenantService_ServiceDesc: service.NewTenantService(),
	}
	grpcServer := grpc_server.NewGRPCServer(certs, model_shared.ModuleCore, serverPort, services)
	if grpcServer == nil {
		return
	}

	// WaitGroup to wait for the gRPC server goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Run gRPC Server
		if err := grpcServer.ListenAndServe(quit); err != nil {
			return
		}
	}()

	// TODO: uncomment when all infra is ready for testing
	// if err := createDefaultData(); err != nil {
	// 	return
	// }
	// Wait for OS signal
	<-stopChan

	// Signal the gRPC server to stop
	close(quit)

	// Wait for the gRPC server goroutine to finish
	wg.Wait()
}

/*

var (
	permissionAllString = fmt.Sprintf("%s:%s", model_auth.ResourceTypeAll, model_auth.PermissionActionAll)
)

func createDefaultData() error {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)

	logger.Debug("Creating default data")

	logger.Debug("Creating system tenant")
	if err := createSystemTenant(mongo.NewBaseCollectionHandler[model_core.Tenant](string(model_mongo.TenantsCollection), logger.NewBaseLogger(model_shared.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system tenant", "error", err)
		return err
	}
	logger.Debug("System tenant created")

	logger.Debug("Creating system admin role")
	if err := createSystemAdminRole(mongo.NewBaseCollectionHandler[model_auth.Role](string(model_mongo.RolesCollection), logger.NewBaseLogger(model_shared.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system admin role", "error", err)
		return err
	}
	logger.Debug("System admin role created")

	logger.Debug("Creating system admin permission")
	if err := createSystemAdminPermission(mongo.NewBaseCollectionHandler[model_auth.Permission](string(model_mongo.PermissionsCollection), logger.NewBaseLogger(model_shared.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system admin permission", "error", err)
		return err
	}
	logger.Debug("System admin permission created")

	logger.Debug("Creating system admin user")
	if err := createSystemAdminUser(mongo.NewBaseCollectionHandler[model_core.User](string(model_mongo.UsersCollection), logger.NewBaseLogger(model_shared.ModuleAuth))); err != nil {
		logger.Fatal("failed to create system admin user", "error", err)
		return err
	}
	logger.Debug("System admin user created")
	return nil
}

func createSystemTenant(collection collection.CollectionHandler[model_core.Tenant]) error {
	tenant := model_core.Tenant{
		Name:      "System",
		Status:    model_auth.TenantStatusActive,
		CreatedBy: "System",
	}
	systemTenantID, err := collection.Create(tenant)
	if err != nil || systemTenantID == "" {
		return infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}
	db.SystemTenantID = systemTenantID
	return nil
}

func createSystemAdminRole(collection collection.CollectionHandler[model_auth.Role]) error {
	role := model_auth.Role{
		Name:        "SystemAdmin",
		Description: "System admin role",
		Permissions: []string{permissionAllString},
	}
	roleID, err := collection.Create(role)
	if err != nil || roleID == "" {
		return infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}
	db.SystemAdminRoleID = roleID
	return nil
}

func createSystemAdminPermission(collection collection.CollectionHandler[model_auth.Permission]) error {

	permission := model_auth.Permission{
		TenantID:         db.SystemTenantID,
		Resource:         model_auth.ResourceTypeAll,
		Action:           model_auth.PermissionActionAll,
		CreatedBy:        "System",
		DisplayName:      "System Controller",
		PermissionString: permissionAllString,
		IsDangerous:      true,
	}
	permissionID, err := collection.Create(permission)
	if err != nil || permissionID == "" {
		return infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}
	db.SystemAdminPermissionID = permissionID
	return nil
}

func createSystemAdminUser(collectionHandler collection.CollectionHandler[model_core.User]) error {
	hash, _ := password.HashPassword(db.SystemAdminPassword)
	user := model_core.User{
		TenantID:     db.SystemTenantID,
		Username:     db.SystemAdminUser,
		Email:        db.SystemAdminEmail,
		PasswordHash: hash,
		Status:       model_auth.UserStatusActive,
		CreatedBy:    "System",
		Roles: []model_core.UserRole{
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
		return infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}
	db.SystemAdminUserID = userID
	return nil
}
*/
