package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"erp.localhost/internal/config/service"
	grpc_server "erp.localhost/internal/infra/grpc/server"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_config "erp.localhost/internal/infra/proto/config/v1"
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

	certs := model_shared.NewCerts()
	if certs == nil {
		return
	}
	services := map[*grpc.ServiceDesc]any{
		&proto_config.ConfigService_ServiceDesc: service.NewConfigService(),
	}
	grpcServer := grpc_server.NewGRPCServer(certs, model_shared.ModuleAuth, serverPort, services)

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

	// Wait for OS signal
	<-stopChan

	// Signal the gRPC server to stop
	close(quit)

	// Wait for the gRPC server goroutine to finish
	wg.Wait()
}
