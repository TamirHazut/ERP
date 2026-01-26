package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"erp.localhost/internal/config/service"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/grpc/server"
	grpc_server "erp.localhost/internal/infra/grpc/server"
	"erp.localhost/internal/infra/logging/logger"
	configv1 "erp.localhost/internal/infra/model/config/v1"
	model_shared "erp.localhost/internal/infra/model/shared"
)

const (
	ServerPort = 5002
)

func Main() {
	logger := logger.NewBaseLogger(model_shared.ModuleConfig)
	defer logger.Close()
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
	// Create server
	logger.Info("Creating gRPC server...")
	srv, err := grpc_server.NewGRPCServer(&server.Config{
		Port:             ServerPort,
		Module:           model_shared.ModuleConfig,
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

	/* Register services */
	logger.Info("Registering gRPC services...")
	configService := service.NewConfigService()
	srv.RegisterService(&configv1.ConfigService_ServiceDesc, configService)

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

	logger.Warn("gRPC server shutdown...")
	// Wait for OS signal
	<-stopChan

	// Signal the gRPC server to stop
	close(quit)

	// Wait for the gRPC server goroutine to finish
	wg.Wait()
	logger.Warn("gRPC server stopped")
}
