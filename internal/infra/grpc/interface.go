package grpc

import "google.golang.org/grpc"

//go:generate mockgen -destination=mock/mock_rpc_client.go -package=mock erp.localhost/internal/infra/grpc/ RPCClient
//go:generate mockgen -destination=mock/mock_rpc_server.go -package=mock erp.localhost/internal/infra/grpc/ RPCServer
type RPCClient interface {
	Conn() *grpc.ClientConn
	Close() error
}

type RPCServer interface {
	ListenAndServe(quit <-chan struct{}) error
}
