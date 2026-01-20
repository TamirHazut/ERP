package server

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/grpc/interceptor"
	"erp.localhost/internal/infra/logging/logger"
	"erp.localhost/internal/infra/model/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

//go:generate mockgen -destination=mock/mock_rpc_server.go -package=mock erp.localhost/internal/infra/grpc/server RPCServer
type RPCServer interface {
	Server() *grpc.Server
	RegisterService(desc *grpc.ServiceDesc, impl interface{})
	ListenAndServe(quit <-chan struct{}) error
}

type Config struct {
	Port              int
	Certs             *shared.Certs
	Module            shared.Module
	Insecure          bool
	EnableReflection  bool
	MaxConnectionIdle time.Duration
	MaxConnectionAge  time.Duration
	KeepAliveTime     time.Duration
	KeepAliveTimeout  time.Duration
}

type GRPCServer struct {
	server *grpc.Server
	config *Config
	logger logger.Logger
}

func NewGRPCServer(config *Config, logger logger.Logger) (*GRPCServer, error) {
	// Build server options
	opts, err := buildServerOptions(config, logger)
	if err != nil {
		logger.Error("failed to build options", "error", err)
		return nil, err
	}

	grpcServer := grpc.NewServer(opts...)

	// Enable reflection if requested
	if config.EnableReflection {
		reflection.Register(grpcServer)
		logger.Info("gRPC reflection enabled")
	}

	return &GRPCServer{
		server: grpcServer,
		config: config,
		logger: logger,
	}, nil
}

// Server returns the underlying grpc.Server for manual service registration
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// RegisterService registers a service implementation with the server
func (s *GRPCServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
	s.logger.Info("registered gRPC service", "service", desc.ServiceName)
}

func (s *GRPCServer) ListenAndServe(quit <-chan struct{}) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		err := infra_error.Internal(infra_error.InternalGRPCError, err)
		s.logger.Error("failed to listen", "error", err)
		return err
	}

	s.logger.Info("gRPC server listening", "port", s.config.Port)

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

	s.logger.Info("initiating graceful shutdown...")
	s.server.GracefulStop()

	<-serverStopped
	s.logger.Info("server shutdown complete")

	return nil
}

func buildServerOptions(config *Config, logger logger.Logger) ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption

	// Add interceptors (from your interceptor package)
	opts = append(opts,
		grpc.ChainUnaryInterceptor(
			// Add your interceptors here
			interceptor.ServerLoggingInterceptor(logger),
		),
	)

	// Keep-alive settings
	if config.KeepAliveTime > 0 {
		opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    config.KeepAliveTime,
			Timeout: config.KeepAliveTimeout,
		}))
	}

	// Connection limits
	if config.MaxConnectionIdle > 0 || config.MaxConnectionAge > 0 {
		opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: config.MaxConnectionIdle,
			MaxConnectionAge:  config.MaxConnectionAge,
		}))
	}

	// Handle credentials
	if config.Insecure {
		logger.Warn("running server in INSECURE mode (no TLS)")
		// No additional credentials needed for insecure
	} else {
		tlsOpts, err := buildTLSOptions(config.Certs)
		if err != nil {
			return nil, err
		}
		opts = append(opts, tlsOpts...)
	}

	return opts, nil
}

func buildTLSOptions(certs *shared.Certs) ([]grpc.ServerOption, error) {
	if certs == nil || !certs.IsValidCerts() {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("invalid or missing certificates"))
	}

	// Load client certificate
	serverCert, err := tls.LoadX509KeyPair(certs.Cert, certs.Key)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("failed to load client certificate")).WithError(err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(certs.CACert)
	if err != nil {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("failed to read CA certificate")).WithError(err)
	}
	// Create cert pool
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("failed to append CA certificate"))
	}

	// Create TLS config for mTLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	creds := credentials.NewTLS(tlsConfig)

	return []grpc.ServerOption{
		grpc.Creds(creds),
	}, nil
}
