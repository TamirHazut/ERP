package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"erp.localhost/internal/infra/logging"
	shared_models "erp.localhost/internal/infra/models/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCClient struct {
	client *grpc.ClientConn
	certs  *shared_models.Certs
	logger *logging.Logger
}

func NewGRPCClient(address string, certs *shared_models.Certs, module shared_models.Module) *GRPCClient {
	logger := logging.NewLogger(module)
	// Build dial options
	opts := getmTLSClientOptions(certs, logger)
	if opts == nil {
		return nil
	}

	// Create connection
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		logger.Error("failed to connect to gRPC server", "address", address, "error", err)
		return nil
	}

	logger.Info("connected to gRPC server", "address", address)

	return &GRPCClient{
		client: conn,
		certs:  certs,
		logger: logger,
	}
}

// Conn returns the underlying connection for creating service clients
func (c *GRPCClient) Conn() *grpc.ClientConn {
	return c.client
}

// Close closes the gRPC connection
func (c *GRPCClient) Close() error {
	if c.client != nil {
		c.logger.Info("closing gRPC client connection")
		return c.client.Close()
	}
	return nil
}

func getmTLSClientOptions(certs *shared_models.Certs, logger *logging.Logger) []grpc.DialOption {
	// If no certs provided, use insecure connection
	if certs == nil || certs.CACert == "" || certs.Cert == "" || certs.Key == "" {
		logger.Error("no certificates provided")
		return nil
	}

	// Load client certificate
	clientCert, err := tls.LoadX509KeyPair(certs.Cert, certs.Key)
	if err != nil {
		logger.Error("failed to load client certificate", "error", err)
		return nil
	}

	// Load CA certificate
	caCert, err := os.ReadFile(certs.CACert)
	if err != nil {
		logger.Error("failed to read CA certificate", "error", err)
		return nil
	}

	// Create cert pool
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		logger.Error("failed to append CA certificate")
		return nil
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}

	creds := credentials.NewTLS(tlsConfig)

	return []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
}
