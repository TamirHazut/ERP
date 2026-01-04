package grpc

import "google.golang.org/grpc"

//go:generate mockgen -destination=mocks/mock_rpc_client.go -package=mocks erp.localhost/internal/infra/grpc/ RPCClient
//go:generate mockgen -destination=mocks/mock_rpc_server.go -package=mocks erp.localhost/internal/infra/grpc/ RPCServer
type RPCClient interface {
	Conn() *grpc.ClientConn
	Close() error
}

type RPCServer interface {
	ListenAndServe(quit <-chan struct{}) error
}
