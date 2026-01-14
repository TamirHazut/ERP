package cmd

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"erp.localhost/internal/auth/api"
	collection_auth "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/auth/rbac"
	"erp.localhost/internal/auth/service"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	"erp.localhost/internal/infra/grpc/server"
	grpc_server "erp.localhost/internal/infra/grpc/server"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
)

const (
	serverPort = 5000
)

// TODO: when breaking to microservices, this will be the entry point for the auth service
func Main() {
	// Channel to listen for OS signals for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to signal the gRPC server goroutine to stop
	quit := make(chan struct{})

	insecure := false
	certs := model_shared.NewCerts()
	if certs == nil {
		insecure = true
	}

	// Create server
	srv, err := grpc_server.NewGRPCServer(&server.Config{
		Port:             serverPort,
		Module:           model_shared.ModuleAuth,
		Insecure:         insecure, // Set to false for production with certs
		Certs:            certs,
		EnableReflection: true,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	})
	if err != nil {
		return
	}
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)
	roleManager := createRoleManager(logger)
	permManager := createPermissionManager(logger)
	verificationManager := createVerificationManager(logger)

	if roleManager == nil || permManager == nil || verificationManager == nil {
		return
	}
	rbacAPI := api.NewRBACAPI(roleManager, permManager, verificationManager, logger)
	authAPI := api.NewAuthAPI(logger)

	/* Register services */
	// Role service
	roleService := service.NewRoleService(rbacAPI.Roles)
	srv.RegisterService(&proto_auth.RoleService_ServiceDesc, roleService)
	// Permission service
	permissionService := service.NewPermissionService(rbacAPI.Permissions)
	srv.RegisterService(&proto_auth.PermissionService_ServiceDesc, permissionService)
	// Verification service
	verificationService := service.NewVerificationService(rbacAPI.Verification)
	srv.RegisterService(&proto_auth.VerificationService_ServiceDesc, verificationService)
	// Auth service
	authService := service.NewAuthService(authAPI)
	srv.RegisterService(&proto_auth.AuthService_ServiceDesc, authService)
	// user service
	userService := service.NewUserService(authAPI, rbacAPI)
	srv.RegisterService(&proto_auth.UserService_ServiceDesc, userService)
	// Tenant service
	tenantService := service.NewTenantService(authAPI, rbacAPI)
	srv.RegisterService(&proto_auth.TenantService_ServiceDesc, tenantService)

	// WaitGroup to wait for the gRPC server goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Run gRPC Server
		if err := srv.ListenAndServe(quit); err != nil {
			return
		}
	}()

	// Wait for OS signal
	<-stopChan

	// Signal the gRPC server to stop
	close(quit)

	// Wait for the gRPC server goroutine to finish
	wg.Wait()
}

func createRoleManager(logger logger.Logger) *rbac.RoleManager {
	roleHandler := mongo_collection.NewBaseCollectionHandler[model_auth.Role](model_mongo.AuthDB, model_mongo.RolesCollection, logger)
	if roleHandler == nil {
		logger.Fatal("failed to init role collection", "error", errors.New("failed to create collection handler"))
	}
	rc := collection_auth.NewRoleCollection(roleHandler, logger)

	if rc == nil {
		return nil
	}

	return rbac.NewRoleManager(rc, logger)
}

func createPermissionManager(logger logger.Logger) *rbac.PermissionManager {
	permHandler := mongo_collection.NewBaseCollectionHandler[model_auth.Permission](model_mongo.AuthDB, model_mongo.PermissionsCollection, logger)
	if permHandler == nil {
		logger.Fatal("failed to init permission collection", "error", errors.New("failed to create collection handler"))
	}
	pc := collection_auth.NewPermissionCollection(permHandler, logger)

	if pc == nil {
		return nil
	}

	return rbac.NewPermissionManager(pc, logger)
}

func createVerificationManager(logger logger.Logger) *rbac.VerificationManager {
	permHandler := mongo_collection.NewBaseCollectionHandler[model_auth.Permission](model_mongo.AuthDB, model_mongo.PermissionsCollection, logger)
	if permHandler == nil {
		logger.Fatal("failed to init permission collection", "error", errors.New("failed to create collection handler"))
	}
	pc := collection_auth.NewPermissionCollection(permHandler, logger)

	roleHandler := mongo_collection.NewBaseCollectionHandler[model_auth.Role](model_mongo.AuthDB, model_mongo.RolesCollection, logger)
	if roleHandler == nil {
		logger.Fatal("failed to init role collection", "error", errors.New("failed to create collection handler"))
	}
	rc := collection_auth.NewRoleCollection(roleHandler, logger)

	userHandler := mongo_collection.NewBaseCollectionHandler[model_auth.User](model_mongo.AuthDB, model_mongo.UsersCollection, logger)
	if userHandler == nil {
		logger.Fatal("failed to init user collection", "error", errors.New("failed to create collection handler"))
	}
	uc := collection_auth.NewUserCollection(userHandler, logger)

	tenantHandler := mongo_collection.NewBaseCollectionHandler[model_auth.Tenant](model_mongo.AuthDB, model_mongo.TenantsCollection, logger)
	if tenantHandler == nil {
		logger.Fatal("failed to init tenant collection", "error", errors.New("failed to create collection handler"))
	}
	tc := collection_auth.NewTenantCollection(tenantHandler, logger)

	if rc == nil || uc == nil || pc == nil || tc == nil {
		return nil
	}

	return rbac.NewVerificationManager(uc, rc, pc, tc, logger)

}
