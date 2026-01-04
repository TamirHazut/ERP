package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	auth_proto "erp.localhost/internal/auth/proto/auth/v1"
	service "erp.localhost/internal/auth/service"
	erp_errors "erp.localhost/internal/infra/errors"
	"erp.localhost/internal/infra/logging"
	shared_models "erp.localhost/internal/infra/models/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server *grpc.Server
	certs  *shared_models.Certs
	port   int
	logger *logging.Logger
}

func NewGRPCServer(certs *shared_models.Certs, module shared_models.Module, port int, services map[*grpc.ServiceDesc]any) *GRPCServer {
	logger := logging.NewLogger(module)

	tlsServerOptions := getmTLSServerOptions(certs, logger)
	if len(tlsServerOptions) == 0 {
		return nil
	}
	grpcServer := grpc.NewServer(tlsServerOptions...)
	for serviceDesc, service := range services {
		grpcServer.RegisterService(serviceDesc, service)
	}
	auth_proto.RegisterAuthServiceServer(grpcServer, service.NewAuthService())
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
		err := erp_errors.Internal(erp_errors.InternalGRPCError, err)
		s.logger.Fatal("failed to listen", "error", err)
		return err
	}
	s.logger.Info(fmt.Sprintf("gRPC server is listening on port: %d", s.port))
	// Channel to signal when the server has shut down
	serverStopped := make(chan struct{})

	go func() {
		if err = s.server.Serve(lis); err != nil {
			err := erp_errors.Internal(erp_errors.InternalGRPCError, err)
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

func getmTLSServerOptions(certs *shared_models.Certs, logger *logging.Logger) []grpc.ServerOption {
	if certs == nil || certs.CACert == "" || certs.Cert == "" || certs.Key == "" {
		logger.Fatal("Failed to ")
		return nil
	}
	serverCert, err := tls.LoadX509KeyPair(certs.Cert, certs.Key)
	if err != nil {
		logger.Fatal("failed to load certificate", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
		return nil
	}

	caCert, err := os.ReadFile(certs.CACert)
	if err != nil {
		logger.Fatal("failed to load certificate", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
		return nil
	}

	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		logger.Fatal("failed to load certificate pool", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
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
