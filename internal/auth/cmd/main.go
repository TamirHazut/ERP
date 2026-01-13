package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"erp.localhost/internal/auth/service"
	"erp.localhost/internal/infra/grpc/server"
	grpc_server "erp.localhost/internal/infra/grpc/server"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
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

	rbacService := service.NewRBACService()
	srv.RegisterService(&proto_auth.RBACService_ServiceDesc, rbacService)
	authService := service.NewAuthService()
	srv.RegisterService(&proto_auth.AuthService_ServiceDesc, authService)
	// Register your services
	userService := service.NewUserService()
	srv.RegisterService(&proto_auth.UserService_ServiceDesc, userService)
	tenantService := service.NewTenantService(nil, nil)
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
