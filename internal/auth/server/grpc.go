package server

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"

	auth_proto "erp.localhost/internal/auth/proto/auth/v1"
	service "erp.localhost/internal/auth/service"
	common_models "erp.localhost/internal/common/models"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	caCert = "../certs/ca.crt"
	cert   = "../certs/cert.pem"
	key    = "../certs/key.pem"
)

func getmTLSServerOptions(logger *logging.Logger) []grpc.ServerOption {
	serverCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		logger.Fatal("failed to load certificate", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
	}

	caCert, err := os.ReadFile(caCert)
	if err != nil {
		logger.Fatal("failed to load certificate", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
	}

	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		logger.Fatal("failed to load certificate pool", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
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

func StartGRPCServer() {
	logger := logging.NewLogger(common_models.ModuleAuth)
	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		logger.Fatal("failed to listen", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
	}

	tlsServerOptions := getmTLSServerOptions(logger)
	grpcServer := grpc.NewServer(tlsServerOptions...)

	auth_proto.RegisterAuthServiceServer(grpcServer, service.NewAuthService())
	reflection.Register(grpcServer)
	if err = grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve", "error", erp_errors.Internal(erp_errors.InternalGRPCError, err))
	}
}
