package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"erp.localhost/internal/infra/grpc/client"
	"erp.localhost/internal/infra/grpc/server"
	grpc_server "erp.localhost/internal/infra/grpc/server"
	"erp.localhost/internal/infra/logging/logger"
	model_shared "erp.localhost/internal/infra/model/shared"
)

const (
	ServerPort = 5001
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

	ctx := context.Background()
	logger := logger.NewBaseLogger(model_shared.ModuleCore)
	// Create RBAC client
	rbacClient, err := client.NewRBACClient(ctx, &client.Config{
		Address:  "localhost:5000",
		Module:   model_shared.ModuleCore,
		Insecure: insecure,
		Certs:    certs,
	}, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer rbacClient.Close()

	// Create server
	srv, err := grpc_server.NewGRPCServer(&server.Config{
		Port:             ServerPort,
		Module:           model_shared.ModuleCore,
		Insecure:         insecure, // Set to false for production with certs
		Certs:            certs,
		EnableReflection: true,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	})
	if err != nil {
		return
	}

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
