package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	configv1 "erp.localhost/internal/infra/proto/config/v1"
	"erp.localhost/internal/config/service"
	infra_grpc "erp.localhost/internal/infra/grpc"
	shared_models "erp.localhost/internal/infra/models/shared"
	"google.golang.org/grpc"
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

	certs := shared_models.NewCerts()
	if certs == nil {
		return
	}
	services := map[*grpc.ServiceDesc]any{
		&configv1.ConfigService_ServiceDesc: service.NewConfigService(),
	}
	grpcServer := infra_grpc.NewGRPCServer(certs, shared_models.ModuleAuth, serverPort, services)

	if grpcServer == nil {
		return
	}

	// WaitGroup to wait for the gRPC server goroutine to finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Run gRPC Server
		grpcServer.ListenAndServe(quit)
	}()

	// Wait for OS signal
	<-stopChan

	// Signal the gRPC server to stop
	close(quit)

	// Wait for the gRPC server goroutine to finish
	wg.Wait()
}
