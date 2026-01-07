package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_shared "erp.localhost/internal/infra/model/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

//go:generate mockgen -destination=mock/mock_rpc_server.go -package=mock erp.localhost/internal/infra/grpc/server RPCServer
type RPCServer interface {
	ListenAndServe(quit <-chan struct{}) error
}

type GRPCServer struct {
	server *grpc.Server
	certs  *model_shared.Certs
	port   int
	logger logger.Logger
}

func NewGRPCServer(certs *model_shared.Certs, module model_shared.Module, port int, services map[*grpc.ServiceDesc]any) *GRPCServer {
	logger := logger.NewBaseLogger(module)

	tlsServerOptions := getmTLSServerOptions(certs, logger)
	if len(tlsServerOptions) == 0 {
		return nil
	}
	grpcServer := grpc.NewServer(tlsServerOptions...)
	for serviceDesc, service := range services {
		grpcServer.RegisterService(serviceDesc, service)
	}
	reflection.Register(grpcServer)
	return &GRPCServer{
		server: grpcServer,
		certs:  certs,
		port:   port,
		logger: logger,
	}

}

func (s *GRPCServer) ListenAndServe(quit <-chan struct{}) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		err := infra_error.Internal(infra_error.InternalGRPCError, err)
		s.logger.Fatal("failed to listen", "error", err)
		return err
	}
	s.logger.Info(fmt.Sprintf("gRPC server is listening on port: %d", s.port))
	// Channel to signal when the server has shut down
	serverStopped := make(chan struct{})

	go func() {
		if err = s.server.Serve(lis); err != nil {
			err := infra_error.Internal(infra_error.InternalGRPCError, err)
			s.logger.Fatal("failed to serve", "error", err)
		}
		close(serverStopped)
	}()

	<-quit

	s.logger.Info("Initiating graceful stop for gRPC server...")
	s.server.GracefulStop()

	<-serverStopped

	return nil
}

func getmTLSServerOptions(certs *model_shared.Certs, logger logger.Logger) []grpc.ServerOption {
	if certs == nil || certs.CACert == "" || certs.Cert == "" || certs.Key == "" {
		logger.Fatal("Failed to ")
		return nil
	}
	serverCert, err := tls.LoadX509KeyPair(certs.Cert, certs.Key)
	if err != nil {
		logger.Fatal("failed to load certificate", "error", infra_error.Internal(infra_error.InternalGRPCError, err))
		return nil
	}

	caCert, err := os.ReadFile(certs.CACert)
	if err != nil {
		logger.Fatal("failed to load certificate", "error", infra_error.Internal(infra_error.InternalGRPCError, err))
		return nil
	}

	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		logger.Fatal("failed to load certificate pool", "error", infra_error.Internal(infra_error.InternalGRPCError, err))
		return nil
	}
	caCertPool.AppendCertsFromPEM([]byte(caCert))
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	credentials := credentials.NewTLS(tlsConfig)
	return []grpc.ServerOption{
		grpc.Creds(credentials),
	}
}
