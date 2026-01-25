package cmd

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"erp.localhost/internal/auth/api"
	"erp.localhost/internal/auth/handler"
	"erp.localhost/internal/auth/rbac"
	"erp.localhost/internal/auth/service"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/grpc/server"
	grpc_server "erp.localhost/internal/infra/grpc/server"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_shared "erp.localhost/internal/infra/model/shared"
)

const (
	ServerPort = 5000
)

// TODO: when breaking to microservices, this will be the entry point for the auth service
func Main() {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)
	logger.Info("Starting service...")
	// Channel to listen for OS signals for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to signal the gRPC server goroutine to stop
	quit := make(chan struct{})

	insecure := false
	certs := model_shared.NewCerts()
	if certs == nil {
		logger.Warn("configuring insecure")
		insecure = true
	}

	// Create server
	logger.Info("Creating gRPC server...")
	srv, err := grpc_server.NewGRPCServer(&server.Config{
		Port:             ServerPort,
		Module:           model_shared.ModuleAuth,
		Insecure:         insecure, // Set to false for production with certs
		Certs:            certs,
		EnableReflection: true,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	}, logger)
	if err != nil {
		logger.Error(infra_error.Internal(infra_error.InternalGRPCError, err).Error())
		return
	}

	roleHanlder := createRoleHandler(logger)
	if roleHanlder == nil {
		logger.Error(infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("failed to create role manager")).Error())
		return
	}
	permHandler := createPermissionHandler(logger)
	if permHandler == nil {
		logger.Error(infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("failed to create permission manager")).Error())
		return
	}
	verificationManager := createVerificationManager(logger)
	if verificationManager == nil {
		logger.Error(infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("failed to create verification manager")).Error())
		return
	}
	rbacAPI := api.NewRBACAPI(roleHanlder, permHandler, verificationManager, logger)
	userAPI, err := api.NewUserAPI(rbacAPI, logger)
	authAPI, err := api.NewAuthAPI(rbacAPI, userAPI, logger)
	tenantAPI, err := api.NewTenantAPI(authAPI, rbacAPI, userAPI, logger)

	/* Register services */
	logger.Info("Registering gRPC services...")
	// Role service
	roleService := service.NewRoleService(rbacAPI.Roles, logger)
	srv.RegisterService(&authv1.RoleService_ServiceDesc, roleService)
	// Permission service
	permissionService := service.NewPermissionService(rbacAPI.Permissions, logger)
	srv.RegisterService(&authv1.PermissionService_ServiceDesc, permissionService)
	// Verification service
	verificationService := service.NewVerificationService(rbacAPI.Verification, logger)
	srv.RegisterService(&authv1.VerificationService_ServiceDesc, verificationService)
	// Auth service
	authService := service.NewAuthService(authAPI, logger)
	srv.RegisterService(&authv1.AuthService_ServiceDesc, authService)
	// user service
	userService := service.NewUserService(userAPI, logger)
	srv.RegisterService(&authv1.UserService_ServiceDesc, userService)
	// Tenant service
	tenantService := service.NewTenantService(tenantAPI, logger)
	srv.RegisterService(&authv1.TenantService_ServiceDesc, tenantService)

	// WaitGroup to wait for the gRPC server goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Run gRPC Server
		if err := srv.ListenAndServe(quit); err != nil {
			logger.Warn("gRPC server stopped", "error", err)
			return
		}
	}()

	// Wait for OS signal
	<-stopChan

	logger.Warn("gRPC server shutdown...")
	// Signal the gRPC server to stop
	close(quit)

	// Wait for the gRPC server goroutine to finish
	wg.Wait()
	logger.Warn("gRPC server stopped")
}

func createRoleHandler(logger logger.Logger) *handler.RoleHandler {
	hanlder, err := handler.NewRoleHandler(logger)
	if err != nil {
		logger.Fatal("failed to init role handler", "error", err)
	}
	return hanlder
}

func createPermissionHandler(logger logger.Logger) *handler.PermissionHandler {
	hanlder, err := handler.NewPermissionHandler(logger)
	if err != nil {
		logger.Fatal("failed to init role handler", "error", err)
	}
	return hanlder
}
func createUserManager(logger logger.Logger) *handler.UserHandler {
	hanlder, err := handler.NewUserHandler(logger)
	if err != nil {
		logger.Fatal("failed to init role handler", "error", err)
	}
	return hanlder
}
func createTenantManager(logger logger.Logger) *handler.TenantHandler {
	hanlder, err := handler.NewTenantHandler(logger)
	if err != nil {
		logger.Fatal("failed to init role handler", "error", err)
	}
	return hanlder
}

func createVerificationManager(logger logger.Logger) *rbac.VerificationManager {
	uh := createUserManager(logger)
	rh := createRoleHandler(logger)
	ph := createPermissionHandler(logger)
	th := createTenantManager(logger)

	if rh == nil || ph == nil || uh == nil || th == nil {
		return nil
	}

	return rbac.NewVerificationManager(uh, rh, ph, th, logger)

}
