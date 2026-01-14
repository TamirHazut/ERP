package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"erp.localhost/internal/config/service"
	"erp.localhost/internal/infra/grpc/server"
	grpc_server "erp.localhost/internal/infra/grpc/server"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_config "erp.localhost/internal/infra/proto/generated/config/v1"
)

const (
	serverPort = 5002
)

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
		Module:           model_shared.ModuleConfig,
		Insecure:         insecure, // Set to false for production with certs
		Certs:            certs,
		EnableReflection: true,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	})
	if err != nil {
		return
	}
	configService := service.NewConfigService()
	srv.RegisterService(&proto_config.ConfigService_ServiceDesc, configService)

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
