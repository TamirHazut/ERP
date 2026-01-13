package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"time"

	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/grpc/interceptor"
	"erp.localhost/internal/infra/logging/logger"
	model_shared "erp.localhost/internal/infra/model/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate mockgen -destination=mock/mock_rpc_client.go -package=mock erp.localhost/internal/infra/grpc/client RPCClient
type RPCClient interface {
	Conn() *grpc.ClientConn
	Close() error
}

type Config struct {
	Address        string
	Certs          *model_shared.Certs
	Module         model_shared.Module
	Insecure       bool
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
}

type GRPCClient struct {
	conn   *grpc.ClientConn
	config *Config
	logger logger.Logger
}

func NewGRPCClient(ctx context.Context, config *Config, logger logger.Logger) (*GRPCClient, error) {
	// Build dial options
	opts, err := buildDialOptions(config, logger)
	if err != nil {
		logger.Error("failed to build options", "error", err)
		return nil, err
	}

	conn, err := grpc.NewClient(config.Address, opts...)
	if err != nil {
		logger.Error("failed to connect to gRPC server", "address", config.Address, "error", err)
		return nil, err
	}

	logger.Info("connected to gRPC server", "address", config.Address)

	return &GRPCClient{
		conn:   conn,
		config: config,
		logger: logger,
	}, nil
}

// Conn returns the underlying connection for creating service clients
func (c *GRPCClient) Conn() *grpc.ClientConn {
	return c.conn
}

// Close closes the gRPC connection
func (c *GRPCClient) Close() error {
	if c.Conn() != nil {
		c.logger.Info("closing gRPC client connection")
		return c.Conn().Close()
	}
	return nil
}

func buildDialOptions(config *Config, logger logger.Logger) ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(
			interceptor.ClientLoggingInterceptor(logger),
			// Add more interceptors as needed
		),
	}

	// Handle credentials
	if config.Insecure {
		logger.Warn("using insecure connection (no TLS)")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsOpts, err := buildTLSOptions(config.Certs)
		if err != nil {
			logger.Error("failed to configure mTLS", "error", err)
			return nil, err
		}
		opts = append(opts, tlsOpts...)
	}

	return opts, nil
}

func buildTLSOptions(certs *model_shared.Certs) ([]grpc.DialOption, error) {
	if certs == nil || !certs.IsValidCerts() {
		return nil, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("invalid or missing certificates"))
	}

	// Load client certificate
	clientCert, err := tls.LoadX509KeyPair(certs.Cert, certs.Key)
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

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}

	creds := credentials.NewTLS(tlsConfig)

	return []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}, nil
}
